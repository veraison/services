// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/veraison/ear"
	"github.com/veraison/parsectpm"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/swid"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	ScopeTrustAnchor = "trust anchor"
	ScopeRefValues   = "ref values"
)

type EvidenceHandler struct{}

type SwAttr struct {
	AlgID   *uint64 `json:"parsec-tpm.alg-id"`
	ClassID *[]byte `json:"parsec-tpm.class-id"`
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
	ClassID  *[]byte `json:"parsec-tpm.class-id"`
	InstID   *string `json:"parsec-tpm.instance-id"`
}

type TaEndorsements struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Attr   TaAttr `json:"attributes"`
}

func (s EvidenceHandler) GetName() string {
	return "parsec-tpm-evidence-handler"
}

func (s EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

func (s EvidenceHandler) SynthKeysFromRefValue(tenantID string, refVals *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts(ScopeRefValues, tenantID, refVals.GetAttributes())
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts(ScopeTrustAnchor, tenantID, ta.GetAttributes())
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var ev parsectpm.Evidence
	err := ev.FromCBOR(token.Data)
	if err != nil {
		return "", handler.BadEvidence(err)
	}

	kat := ev.Kat
	if kat == nil {
		return "", errors.New("no key attestation token to fetch Key ID")
	}
	kid := *kat.KID
	instance_id := base64.StdEncoding.EncodeToString(kid)
	return parsecTpmLookupKey(ScopeTrustAnchor, token.TenantId, "", instance_id), nil

}

func (s EvidenceHandler) ExtractClaims(token *proto.AttestationToken, trustAnchor string) (*handler.ExtractedClaims, error) {
	var (
		evidence    parsectpm.Evidence
		endorsement TaEndorsements
		extracted   handler.ExtractedClaims
	)

	err := evidence.FromCBOR(token.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	claimsSet, err := evidenceAsMap(evidence)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	extracted.ClaimsSet = claimsSet
	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		log.Error("Could not decode Endorsements in ExtractClaims: %w", err)
		return nil, fmt.Errorf("could not decode endorsement: %w", err)
	}

	class_id := base64.StdEncoding.EncodeToString(*endorsement.Attr.ClassID)
	extracted.ReferenceID = parsecTpmLookupKey(ScopeRefValues, token.TenantId, class_id, "")
	return &extracted, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(token *proto.AttestationToken, trustAnchor string, endorsements []string) error {
	var endorsement TaEndorsements

	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		log.Error("Could not decode Endorsement in ValidateEvidenceIntegrity: %w", err)
		return fmt.Errorf("could not decode endorsement: %w", err)
	}
	ta := *endorsement.Attr.VerifKey
	block, rest := pem.Decode([]byte(ta))

	if block == nil {
		log.Error("Could not get TA PEM Block ValidateEvidenceIntegrity")
		return errors.New("could not extract trust anchor PEM block")
	}

	if len(rest) != 0 {
		return errors.New("trailing data found after PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return fmt.Errorf("unsupported key type: %q", block.Type)
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse public key: %w", err)
	}

	var ev parsectpm.Evidence
	if err = ev.FromCBOR(token.Data); err != nil {
		return handler.BadEvidence(err)
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

func synthKeysFromParts(scope, tenantID string, parts *structpb.Struct) ([]string, error) {
	var (
		instance string
		class    string
		fields   map[string]*structpb.Value
		err      error
	)

	fields, err = common.GetFieldsFromParts(parts)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	if scope == ScopeTrustAnchor {
		instance, err = common.GetMandatoryPathSegment("parsec-tpm.instance-id", fields)
		if err != nil {
			return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
		}
	}

	class, err = common.GetMandatoryPathSegment("parsec-tpm.class-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}
	return []string{parsecTpmLookupKey(scope, tenantID, class, instance)}, nil
}

func parsecTpmLookupKey(scope, tenantID, class, instance string) string {
	var absPath []string

	switch scope {
	case ScopeTrustAnchor:
		absPath = []string{instance}
	case ScopeRefValues:
		absPath = []string{class}
	}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func evidenceAsMap(e parsectpm.Evidence) (map[string]interface{}, error) {
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
		log.Error("match PCR failed: %w", err)
		return fmt.Errorf("match PCR failed: %w", err)
	}

	if err := matchPCRDigest(pcrDigest, hashAlgID, eds); err != nil {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Error("match PCR Digest failed: %w", err)
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
	kd, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("unable to marshal public key: %w", err)
	}
	pubkey := base64.StdEncoding.EncodeToString(kd)
	appraisal.AppraisalExtensions.VeraisonKeyAttestation = &map[string]interface{}{
		"akpub": pubkey,
	}
	appraisal.UpdateStatusFromTrustVector()
	appraisal.VeraisonAnnotatedEvidence = &evidence
	return nil
}

func mapAsEvidence(in map[string]interface{}) (*parsectpm.Evidence, error) {
	evidence := &parsectpm.Evidence{}
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
				log.Error("malformed endorsements: %v", end)
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
// within parsectpm library
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
