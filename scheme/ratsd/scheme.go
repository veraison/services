// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ratsd

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

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

	return nil, nil
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	return extractClaims(evidence.Data)
}

// TO DO We MUST do Validate Evidence Integrity, but for NOW, the RatsD outer token, does not have integrity.
// Kept as a place holder to revisit again, once https://github.com/veraison/ratsd/issues/65 is fully implemented.
// For now this Method, just checks that the `nonce` from Session matches the `nonce` inside the Evidence(token)
func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	eat, err := extractClaims(evidence.Data)
	if err != nil {
		return handler.BadEvidence(err)
	}
	evNonce := eat["eat_nonce"].([]byte)
	if !bytes.Equal(evNonce, evidence.Nonce) {
		return handler.BadEvidence(
			"freshness: evidence challenge (%s) does not match session nonce (%s)",
			hex.EncodeToString(evNonce),
			hex.EncodeToString(evidence.Nonce),
		)
	}
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
	eat := make(map[string]any)

	err := json.Unmarshal(data, &eat)
	if err != nil {
		return nil, err
	}

	cmwBase64, ok := eat["cmw"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to extract cmw from the evidence: %w", err)
	}
	cmwCbor, err := base64.StdEncoding.DecodeString(cmwBase64)
	if err != nil {
		return nil, err
	}
	eat["cmw"] = cmwCbor

	_, ok = eat["eat_profile"]
	if !ok {
		return nil, errors.New("unable to get eat profile from evidence")
	}

	_, ok = eat["eat_nonce"]
	if !ok {
		return nil, errors.New("unable to get eat nonce from evidence")
	}
	return eat, nil
}
