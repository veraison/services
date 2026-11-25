// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/plugin"
)

// ImplementationWrapper provides an implementation of IHandler
// based on a SchemeDescriptor and a IImplementation implementation.
// The wrapper provides default implementations of some of the IHandler
// methods, and makes some others options. This saves attestations schemes from
// needing to implement the full IHandler interface.
type ImplementationWrapper struct {
	Desc Descriptor
	Impl Implementation
}

func NewSchemeImplementationWrapper(
	desc Descriptor,
	impl Implementation,
) (*ImplementationWrapper, error) {
	if err := desc.Validate(); err != nil {
		return nil, err
	}

	return &ImplementationWrapper{Desc: desc, Impl: impl}, nil
}

func MustNewSchemeImplementationWrapper(
	desc Descriptor,
	impl Implementation,
) *ImplementationWrapper {
	ret, err := NewSchemeImplementationWrapper(desc, impl)
	if err != nil {
		panic(err)
	}

	return ret
}

func (o *ImplementationWrapper) Init(params *plugin.Parameters) error {
	pluginImpl, ok := o.Impl.(interface {
		Init(*plugin.Parameters) error
	})
	if ok {
		return pluginImpl.Init(params)
	}

	return nil
}

func (o *ImplementationWrapper) Fini() error {
	if pluginImpl, ok := o.Impl.(interface {
		Fini() error
	}); ok {
		return pluginImpl.Fini()
	}
	return nil
}

func (o *ImplementationWrapper) GetName() string {
	return PluginNameFromScheme(o.Desc.Name)
}

func (o *ImplementationWrapper) GetAttestationScheme() string {
	return o.Desc.Name
}

func (o *ImplementationWrapper) GetSupportedCorimProfiles() []string {
	return o.Desc.CorimProfiles
}

func (o *ImplementationWrapper) GetSupportedEvidenceMediaTypes() []string {
	return o.Desc.EvidenceMediaTypes
}

func (o *ImplementationWrapper) ValidateCorim(uc *corim.UnsignedCorim) (*ValidateCorimResponse, error) {
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
			if tag.Number != corim.ComidTag {
				continue
			}

			var c comid.Comid

			err := c.FromCBOR(tag.Content)
			if err != nil {
				return nil, fmt.Errorf("decoding failed for CoMID at index %d: %w", i, err)
			}

			err = comidImpl.ValidateComid(&c)
			if err != nil {
				return &ValidateCorimResponse{
					IsValid: false,
					Message: fmt.Sprintf("CoMID at index %d: %s", i, err.Error()),
				}, nil
			}

		}

		return &ValidateCorimResponse{IsValid: true, Message: "<all CoMIDs validated>"}, nil
	}

	return &ValidateCorimResponse{IsValid: true, Message: "<no scheme validation>"}, nil
}

func (o *ImplementationWrapper) GetTrustAnchorIDs(
	evidence *Evidence,
) ([]*comid.Environment, error) {
	impl, ok := o.Impl.(interface {
		GetTrustAnchorIDs(*Evidence) ([]*comid.Environment, error)
	})

	if ok {
		return impl.GetTrustAnchorIDs(evidence)
	}

	return nil, nil
}

func (o *ImplementationWrapper) GetReferenceValueIDs(
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

func (o *ImplementationWrapper) ExtractClaims(
	evidence *Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	if !o.Desc.EvidenceIsSupported(evidence) {
		return nil, fault.BadEvidence("wrong media type: expect %q, but found %q",
			strings.Join(o.Desc.EvidenceMediaTypes, ", "),
			evidence.MediaType,
		)
	}

	return o.Impl.ExtractClaims(evidence, trustAnchors)
}

func (o *ImplementationWrapper) ValidateEvidenceIntegrity(
	evidence *Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	impl, ok := o.Impl.(interface {
		ValidateEvidenceIntegrity(*Evidence, []*comid.KeyTriple, []*comid.ValueTriple) error
	})

	if ok {
		return impl.ValidateEvidenceIntegrity(evidence, trustAnchors, endorsements)
	}

	return nil
}

func (o *ImplementationWrapper) AppraiseClaims(
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
