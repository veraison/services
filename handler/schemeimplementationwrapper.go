// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/vts/appraisal"
)

type SchemeImplementationWrapper struct {
	Desc SchemeDescriptor
	Impl ISchemeImplementation
}

func NewSchemeImplementationWrapper(
	desc SchemeDescriptor,
	impl ISchemeImplementation,
) (*SchemeImplementationWrapper, error) {
	if err := desc.Validate(); err != nil {
		return nil, err
	}

	return &SchemeImplementationWrapper{ Desc: desc, Impl: impl }, nil
}

func MustNewSchemeImplementationWrapper(
	desc SchemeDescriptor,
	impl ISchemeImplementation,
) *SchemeImplementationWrapper {
	ret, err := NewSchemeImplementationWrapper(desc, impl)
	if err != nil {
		panic(err)
	}

	return ret
}

func (o *SchemeImplementationWrapper) GetName() string {
	name := strings.ToLower(strings.ReplaceAll(o.Desc.Name, " ", "-"))
	return fmt.Sprintf("%s-scheme-plugin", name)
}

func (o *SchemeImplementationWrapper) GetAttestationScheme() string {
	return o.Desc.Name
}

func (o *SchemeImplementationWrapper) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{
		"provisioning": o.GetSupportedProvisioningMediaTypes(),
		"verification": o.GetSupportedVerificationMediaTypes(),
	}
}

func (o *SchemeImplementationWrapper) GetSupportedProvisioningMediaTypes() []string {
	ret := make([]string, 0, len(o.Desc.CorimProfiles)*2)

	for _, profile := range o.Desc.CorimProfiles {
		ret = append(ret,
			fmt.Sprintf(`application/rim+cbor; profile="%s"`, profile),
			fmt.Sprintf(`application/rim+cose; profile="%s"`, profile),
		)
	}

	return ret
}

func (o *SchemeImplementationWrapper) GetSupportedVerificationMediaTypes() []string {
	return o.Desc.EvidenceMediaTypes
}

func (o *SchemeImplementationWrapper) ValidateCorim(uc *corim.UnsignedCorim) (*ValidateCorimResponse, error) {
	corimImpl, ok := o.Impl.(interface {
		ValidateCorim(*corim.UnsignedCorim) (*ValidateCorimResponse, error)
	})
	if ok {
		return corimImpl.ValidateCorim(uc)
	}

	comidImpl, ok := o.Impl.(interface {
		ValidateComid(*comid.Comid) error
	})
	if ok {
		for i, tag := range uc.Tags {
			if tag.Number == corim.ComidTag {
				var c comid.Comid

				err := c.FromCBOR(tag.Content)
				if err != nil {
					return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
				}

				err = comidImpl.ValidateComid(&c)
				if err != nil {
					return &ValidateCorimResponse{
						IsValid: true,
						Message: fmt.Sprintf("CoMID at index %d: %s", i, err.Error()),
					}, nil
				}
			}
		}

		return &ValidateCorimResponse{IsValid: true, Message: "<all CoMIDs validated>"}, nil
	}

	return &ValidateCorimResponse{IsValid: true, Message: "<no scheme validation>"}, nil
}

func (o *SchemeImplementationWrapper) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {
	return o.Impl.GetTrustAnchorIDs(evidence)
}

func (o *SchemeImplementationWrapper) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	impl, ok := o.Impl.(interface {
		GetReferenceValueIDs([]*comid.KeyTriple, map[string]any) ([]*comid.Environment, error)
	})

	if ok {
		return impl.GetReferenceValueIDs(trustAnchors, claims)
	}

	return nil, nil
}

func (o *SchemeImplementationWrapper) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	if !o.Desc.EvidenceIsSupported(evidence) {
		return nil, BadEvidence("wrong media type: expect %q, but found %q",
			strings.Join(o.Desc.EvidenceMediaTypes, ", "),
			evidence.MediaType,
		)
	}

	return o.Impl.ExtractClaims(evidence, trustAnchors)
}

func (o *SchemeImplementationWrapper) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	impl, ok := o.Impl.(interface {
		ValidateEvidenceIntegrity(*appraisal.Evidence, []*comid.KeyTriple, []*comid.ValueTriple) error
	})

	if ok {
		return impl.ValidateEvidenceIntegrity(evidence, trustAnchors, endorsements)
	}

	return nil
}

func (o *SchemeImplementationWrapper) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	return o.Impl.AppraiseClaims(claims, endorsements)
}

type ValidateCorimResponse struct {
	IsValid bool   `json:"is-valid"`
	Message string `json:"mesage"`
}

func (o *ValidateCorimResponse) Error() error {
	if o.Message == "" {
		return nil
	}

	return errors.New(o.Message)
}

