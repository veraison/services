// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ratsd

import (
	"errors"

	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	ratsd "github.com/veraison/ratsd/ratsd-token-v2"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/vts/appraisal"
	"go.uber.org/zap"
)

var Descriptor = handler.SchemeDescriptor{
	Name:          "RATSD",
	VersionMajor:  1,
	VersionMinor:  0,
	CorimProfiles: []string{""},
	EvidenceMediaTypes: []string{
		`application/eat+ujcs;; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`,
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

	cmw := ev.GetCollection()

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

	profile, ok := claims["eat_profile"].(string)
	if !ok {
		return nil, errors.New("unable to get eat profile from evidence")
	}
	found := false
	for _, p := range Descriptor.EvidenceMediaTypes {
		if p == profile {
			found = true
			break
		}
	}
	if !found {
		return result, handler.BadEvidence(errors.New("invalid profile in the evidence"))
	}

	// Ratsd Lead Attester has no claims of its own
	return result, nil
}

func extractClaims(data []byte) (map[string]any, error) {
	// extract individual tokens and Lead Attester Token
	// Flatten Out ratsd claims
	return eat, nil
}
