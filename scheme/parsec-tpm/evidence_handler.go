// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/veraison/ear"
	"github.com/veraison/parsec/tpm"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/swid"
)

type EvidenceHandler struct{}

type SwAttr struct {
	AlgID   *uint64 `json:"parsec-tpm.alg-id"`
	ClassID *string `json:"parsec-tpm.class-id"`
	Digest  *[]byte `json:"parsec-tpm.digest"`
	PCR     *uint   `json:"parsec-tpm.pcr"`
}

type Endorsements struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Attr   SwAttr `json:"attributes"`
}

type TaAttr struct {
	VerifKey *string `json:"parsec-tpm.ak-pub"`
	ClassID  *string `json:"parsec-tpm.class-id"`
	InstID   *string `json:"parsec-tpm.instance-id"`
}

type TaEndorsements struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Attr   TaAttr `json:"attributes"`
}

func (s EvidenceHandler) GetName() string {
	return EvidenceHandlerName
}

func (s EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

func (s EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchors []string,
) (map[string]interface{}, error) {
	var evidence tpm.Evidence

	err := evidence.FromCBOR(token.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claims, err := evidenceAsMap(evidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	return claims, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(token *proto.AttestationToken, trustAnchors []string, endorsements []string) error {
	var (
		endorsement TaEndorsements
		ev          tpm.Evidence
	)

	if err := ev.FromCBOR(token.Data); err != nil {
		return handler.BadEvidence(err)
	}

	if err := json.Unmarshal([]byte(trustAnchors[0]), &endorsement); err != nil {
		log.Errorf("Could not decode trust anchor in ValidateEvidenceIntegrity: %v", err)
		return fmt.Errorf("could not decode trust anchor: %w", err)
	}

	ta := *endorsement.Attr.VerifKey
	pk, err := common.DecodePemSubjectPubKeyInfo([]byte(ta))
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchor: %w", err)
	}

	if err := ev.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}

	log.Debug("Token Signature Verified")
	return nil
}

func (s EvidenceHandler) AppraiseEvidence(ec *proto.EvidenceContext, endorsementStrings []string) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(SchemeName)
	var endorsements []Endorsements // nolint:prealloc

	for i, e := range endorsementStrings {
		var endorsement Endorsements

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index: %d, %w", i, err)
		}

		endorsements = append(endorsements, endorsement)
	}
	err := populateAttestationResult(result, ec.Evidence.AsMap(), endorsements)
	return result, err
}

func evidenceAsMap(e tpm.Evidence) (map[string]interface{}, error) {
	data, err := e.ToJSON()
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	return out, err
}

func populateAttestationResult(
	result *ear.AttestationResult,
	evidence map[string]interface{},
	endorsements []Endorsements,
) error {
	appraisal := result.Submods[SchemeName]

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	ev, err := mapAsEvidence(evidence)
	if err != nil {
		return handler.BadEvidence(err)
	}

	attInfo, err := ev.Pat.GetAttestationInfo()
	if err != nil {
		return handler.BadEvidence(err)
	}

	pcrs := attInfo.PCR.PCRinfo.PCRs
	hashAlgID := attInfo.PCR.PCRinfo.HashAlgID
	pcrDigest := attInfo.PCR.PCRDigest

	// Match the Evidence PCR against the endorsements to generate a subset
	// of matching endorsements
	eds, err := matchPCRs(pcrs, hashAlgID, endorsements)
	if err != nil {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Errorf("match PCR failed: %v", err)
		return fmt.Errorf("match PCR failed: %w", err)
	}

	if err := matchPCRDigest(pcrDigest, hashAlgID, eds); err != nil {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Errorf("match PCR Digest failed: %v", err)
		return fmt.Errorf("match failed for PCR Digest: %w", err)
	}

	appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
	appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
	log.Debug("matchPCRs and matchPCRDigest Success")

	// Populate Veraison Key Attestation Extension
	key, err := ev.Kat.DecodePubArea()
	if err != nil {
		return handler.BadEvidence(err)
	}
	if err := appraisal.SetKeyAttestation(key); err != nil {
		return fmt.Errorf("setting extracted public key: %w", err)
	}
	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &evidence
	return nil
}

func mapAsEvidence(in map[string]interface{}) (*tpm.Evidence, error) {
	evidence := &tpm.Evidence{}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	err = evidence.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unable to map to evidence: %w", err)
	}
	return evidence, err
}

// match the evidence PCR's against the received Endorsements
func matchPCRs(pcrs []int, algID uint64, endorsements []Endorsements) ([]Endorsements, error) {
	var eds []Endorsements

	if len(pcrs) == 0 {
		return nil, errors.New("no evidence pcrs to match")
	}

	// Sort the PCRs first
	sort.Ints(pcrs)
	for i, pcr := range pcrs {
		matched := false
		for _, end := range endorsements {
			if (end.Attr.PCR == nil) || (end.Attr.AlgID == nil) {
				log.Errorf("malformed endorsements: %v", end)
				continue
			}

			if (pcr == int(*end.Attr.PCR)) && (algID == *end.Attr.AlgID) {
				eds = append(eds, end)
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("unmatched pcr value: %d at index: %d", pcr, i)
		}
	}
	return eds, nil
}

func matchPCRDigest(pDigest []byte, algID uint64, eds []Endorsements) error {
	// concatenate endorsement PCR Digests to get the resulting hash data
	hdata, err := concatHash(eds)
	if err != nil {
		return fmt.Errorf("hash concatenation failed: %w", err)
	}
	// Compute the hash of resulting hash data
	endHash, err := computeHash(hdata, algID)
	if err != nil {
		return fmt.Errorf("unable to compute digest: %w", err)
	}
	if !bytes.Equal(pDigest, endHash) {
		return errors.New("PCR Digest and Endorsement Digest match failed")
	}
	return nil
}

func concatHash(endorsements []Endorsements) ([]byte, error) {
	var digest []byte
	if len(endorsements) == 0 {
		return nil, errors.New("no endorsments to hash")
	}

	for _, ed := range endorsements {
		digest = append(digest, *ed.Attr.Digest...)
	}
	return digest, nil
}

// hashFunc returns the hash associated with the algorithms supported
// within tpm library
func hashFunc(alg uint64) crypto.Hash {
	switch alg {
	case swid.Sha256:
		return crypto.SHA256
	case swid.Sha384:
		return crypto.SHA384
	case swid.Sha512:
		return crypto.SHA512
	default:
		return 0
	}
}

func computeHash(in []byte, algID uint64) ([]byte, error) {
	h := hashFunc(algID)
	if !h.Available() {
		return nil, fmt.Errorf("unavailable hash function for algID: %d", algID)
	}
	hh := h.New()
	if _, err := hh.Write(in); err != nil {
		return nil, err
	}
	return hh.Sum(nil), nil
}
