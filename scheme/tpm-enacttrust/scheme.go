// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"bytes"
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/vts/appraisal"
)

var Descriptor = handler.SchemeDescriptor{
	Name:         "TPM_ENACTTRUST",
	VersionMajor: 1,
	VersionMinor: 0,
	CorimProfiles: []string{
		ProfileString,
	},
	EvidenceMediaTypes: []string{
		"application/vnd.enacttrust.tpm-evidence",
	},
}

type Implementation struct{}

func NewImplementation() *Implementation {
	return &Implementation{}
}

func (o *Implementation) GetTrustAnchorIDs(
	evidence *appraisal.Evidence,
) ([]*comid.Environment, error) {
	return extractEnvinromentsFromEvidence(evidence)
}

func (o *Implementation) GetReferenceValueIDs(
	trustAnchors []*comid.KeyTriple,
	claims map[string]any,
) ([]*comid.Environment, error) {
	return extractEnvinromentsFromClaims(claims)
}

func (o *Implementation) ExtractClaims(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
) (map[string]any, error) {
	var decoded Token

	if err := decoded.Decode(evidence.Data); err != nil {
		return nil, handler.BadEvidence("could not decode token: %w", err)
	}

	if decoded.AttestationData.Type != tpm2.TagAttestQuote {
		return nil, handler.BadEvidence("wrong TPMS_ATTEST type: want %d, got %d",
			tpm2.TagAttestQuote, decoded.AttestationData.Type)
	}

	pcrs := make([]int64, 0, len(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.PCRs))
	for _, pcr := range decoded.AttestationData.AttestedQuoteInfo.PCRSelection.PCRs {
		pcrs = append(pcrs, int64(pcr))
	}

	claims := make(map[string]any)
	claims["pcr-selection"] = pcrs
	claims["hash-algorithm"] = int64(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.Hash)
	claims["firmware-version"] = decoded.AttestationData.FirmwareVersion
	claims["node-id"] = decoded.NodeId.String()
	claims["pcr-digest"] = []byte(decoded.AttestationData.AttestedQuoteInfo.PCRDigest)

	return claims, nil
}

func (o *Implementation) ValidateEvidenceIntegrity(
	evidence *appraisal.Evidence,
	trustAnchors []*comid.KeyTriple,
	endorsements []*comid.ValueTriple,
) error {
	var decoded Token
	if err := decoded.Decode(evidence.Data); err != nil {
		return handler.BadEvidence("could not decode evidence: %w", err)
	}

	taLen := len(trustAnchors)
	if taLen != 1 {
		return fmt.Errorf("expected exactly one trust anchor; found %d", taLen)
	}

	pubKey, err := extractKey(trustAnchors[0].VerifKeys)
	if err != nil {
		return err
	}

	if err = decoded.VerifySignature(pubKey); err != nil {
		return handler.BadEvidence("could not verify evidence signature: %w", err)
	}

	return nil
}

func (o *Implementation) AppraiseClaims(
	claims map[string]any,
	endorsements []*comid.ValueTriple,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(Descriptor.Name)

	evidenceDigest, ok := claims["pcr-digest"]
	if !ok {
		err := handler.BadEvidence("claims do not contain %q entry", "pcr-digest")
		return result, err
	}

	evidenceDigestBytes, ok := evidenceDigest.([]byte)
	if !ok {
		err := handler.BadEvidence("pcr-digest: expected []byte, found %T", evidenceDigest)
		return result, err
	}

	appraisal := result.Submods[Descriptor.Name]
	appraisal.VeraisonAnnotatedEvidence = &claims

	var endorsedDigests [][]byte // nolint:prealloc
	for i, triple := range endorsements {
		digest, err := extractEndorsedDigest(triple.Measurements.Values)
		if err != nil {
			return nil, fmt.Errorf("endorsement %d: %w", i, err)
		}

		endorsedDigests = append(endorsedDigests, digest)
	}

	matched := false
	for _, endorsedDigest := range endorsedDigests {
		if bytes.Equal(endorsedDigest, evidenceDigestBytes) {
			appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
			*appraisal.Status = ear.TrustTierAffirming

			matched = true
			break
		}
	}

	if !matched {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		*appraisal.Status = ear.TrustTierContraindicated
	}

	return result, nil
}

func extractEnvinromentsFromEvidence(evidence *appraisal.Evidence) ([]*comid.Environment, error) {
	var decoded Token

	if err := decoded.Decode(evidence.Data); err != nil {
		return nil, handler.BadEvidence("could not decode token: %w", err)
	}

	return nodeIDToEnvironments(decoded.NodeId)
}

func extractEnvinromentsFromClaims(claims map[string]any) ([]*comid.Environment, error) {
	nodeID, ok := claims["node-id"]
	if !ok {
		return nil, handler.BadEvidence("no node ID in claims")
	}

	return nodeIDToEnvironments(nodeID)
}

func nodeIDToEnvironments(nodeID any) ([]*comid.Environment, error) {
	instance, err := comid.NewUUIDInstance(nodeID)
	if err != nil {
		return nil, handler.BadEvidence("could not create CoMID instance form node ID: %w", err)
	}

	return []*comid.Environment{
		&comid.Environment{Instance: instance},
	}, nil
}
