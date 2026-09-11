// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ratsd

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	ratsd "github.com/veraison/ratsd/ratsd-token-v2"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/vts/appraisal"
	"github.com/veraison/services/vts/compositeevidenceparser"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:          "RATSD",
	VersionMajor:  1,
	VersionMinor:  0,
	CorimProfiles: []string{""},
	EvidenceMediaTypes: []string{
		`"application/eat-ucs+cbor; eat_profile="tag:github.com,2026:veraison/ratsd/v2"`,
	},
}

type Implementation struct {
	logger *zap.SugaredLogger
}

func NewImplementation() *Implementation {
	return &Implementation{
		logger: log.Named(Descriptor.Name),
	}
}

func (o *Implementation) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {

	/*
			ev := &ratsd.Evidence{}
			ev.UnmarshalCBOR(evidence.Data)

			c := ev.GetClaims()
			oem := c.GetOEMID()
			vendor := FormatInt(oem, 10)
			model := c.GetSWName()

		   Set Vendor and Model in the comid.Environment


	*/
	return nil, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {

	// Here we need to just extract the RatsD token...
	ev := &ratsd.Evidence{}
	ev.UnmarshalCBOR(evidence.Data)

	c, err := ev.GetCollection()
	if err != nil {
		return nil, err
	}
	// The following code can be improved slightly
	cmw, err := c.MarshalCBOR()
	if err != nil {
		return nil, err
	}

	p, err := compositeevidenceparser.GetCEParserFromMediaType(evidence.MediaType)
	if err != nil {
		return nil, fmt.Errorf("unable to fecth parser from received MediaType: %s, %w", evidence.MediaType, err)
	}

	evs, err := p.Parse(cmw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Composite Evidence for the MediaType: %s, %w", evidence.MediaType, err)
	}
	if len(evs) == 0 {
		return nil, errors.New("no data in Evidence")
	}
	claims := make(map[string]any)

	for _, ev := range evs {
		claims[ev.GetMediaType()] = ev
	}
	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	// Here we need to just extract the RatsD token...
	ev := &ratsd.Evidence{}
	ev.UnmarshalCBOR(evidence.Data)
	// Get the Signing Certs from Evidence
	// Get the Intermediate Certs from the Evidence

	// Verify against the Trust Anchors

	// From the trustAnchors from the corim-store, assumption is that comid.KeyTriple will contain the RatsD Trust Root x.509 certificatye
	// anchor := trustAnchors.GetRootX.509Cert
	/*
		  var external []byte
		  // TO DO set the Options correctly, else Verification will fail
			chain, err := ev.message.VerifyWithX5Chain(external, anchors, opts)
			if err != nil{
			// Handle error here
			}


	*/

	/*
		   c :=  ev.GetClaims ()
		Get the Nonce Claim
		Compare the Nonces
	*/

	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)

	mt := Descriptor.EvidenceMediaTypes[0]

	e, ok := claims[mt]
	if !ok {
		return result, fmt.Errorf("unable to extract RATSD Evidence from claims map")
	}

	ev := e.(compositeevidenceparser.ComponentEvidence)
	var c ratsd.Claims
	if err := c.UnmarshalCBOR(ev.GetevidenceData()); err != nil {
		return result, fmt.Errorf("unable to unmarshal claims from RATSD Evidence: %w", err)
	}

	// Appraise All claims by comparing it with ValueTriple for RATSD
	// Note we need to define the Reference Values for RATSD Claims

	// Temporary Hack, Once the RatsD Claims have been appraised, remove it from Claims Map
	// Once EAR Overall Appraisal Status is fixed, then we can check from Overall Appraisal in VTS and skip the loop for RATSD Evidence Verification
	// More robust strategy is required to filter Verification of Lead Attester.

	// For now, Once the RatsD Claims have been appraised, remove it from Claims Map
	delete(claims, mt)
	return result, nil
}
