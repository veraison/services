// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ratsd

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
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
		`application/cmw-collection+cbor; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`,
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

	return nil, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	return nil, nil // TODO
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {

	return nil // TODO
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)

	// For Lead Attester, there is no real claims to Appraise,  as there are no LA Claims
	return result, nil
}
