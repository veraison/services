// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	tpm2 "github.com/google/go-tpm/tpm2"

	"github.com/veraison/ear"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

type EvidenceHandler struct{}

func (s EvidenceHandler) GetName() string {
	return "tpm-enacttrust-evidence-handler"
}

func (s EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

func (s EvidenceHandler) SynthKeysFromRefValue(
	tenantID string,
	swComp *handler.Endorsement,
) ([]string, error) {
	return synthKeysFromAttrs("software component", tenantID, swComp.Attributes)
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return synthKeysFromAttrs("trust anchor", tenantID, ta.Attributes)
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	supported := false
	for _, mt := range EvidenceMediaTypes {
		if token.MediaType == mt {
			supported = true
			break
		}
	}

	if !supported {
		err := handler.BadEvidence(
			"wrong media type: expect %q, but found %q",
			strings.Join(EvidenceMediaTypes, ", "),
			token.MediaType,
		)
		return "", err
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return "", handler.BadEvidence(err)
	}

	return tpmEnactTrustLookupKey(token.TenantId, decoded.NodeId.String()), nil
}

func (s EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchor string,
) (*handler.ExtractedClaims, error) {
	supported := false
	for _, mt := range EvidenceMediaTypes {
		if token.MediaType == mt {
			supported = true
			break
		}
	}

	if !supported {
		return nil, handler.BadEvidence("wrong media type: expect %q, but found %q",
			strings.Join(EvidenceMediaTypes, ", "),
			token.MediaType,
		)
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return nil, handler.BadEvidence("could not decode token: %w", err)
	}

	if decoded.AttestationData.Type != tpm2.TagAttestQuote {
		return nil, handler.BadEvidence("wrong TPMS_ATTEST type: want %d, got %d",
			tpm2.TagAttestQuote, decoded.AttestationData.Type)
	}

	var pcrs []interface{} // nolint:prealloc
	for _, pcr := range decoded.AttestationData.AttestedQuoteInfo.PCRSelection.PCRs {
		pcrs = append(pcrs, int64(pcr))
	}

	evidence := handler.NewExtractedClaims()
	evidence.ClaimsSet["pcr-selection"] = pcrs
	evidence.ClaimsSet["hash-algorithm"] = int64(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.Hash)
	evidence.ClaimsSet["firmware-version"] = decoded.AttestationData.FirmwareVersion
	evidence.ClaimsSet["node-id"] = decoded.NodeId.String()
	evidence.ClaimsSet["pcr-digest"] = []byte(decoded.AttestationData.AttestedQuoteInfo.PCRDigest)
	evidence.ReferenceID = tpmEnactTrustLookupKey(token.TenantId, decoded.NodeId.String())

	return evidence, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchor string,
	endorsements []string,
) error {
	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return handler.BadEvidence("could not decode token: %w", err)
	}

	pubKey, err := parseKey(trustAnchor)
	if err != nil {
		return fmt.Errorf("could not parse trust anchor: %w", err)
	}

	if err = decoded.VerifySignature(pubKey); err != nil {
		return handler.BadEvidence("could not verify token signature: %w", err)
	}

	return nil
}

func (s EvidenceHandler) AppraiseEvidence(
	ec *proto.EvidenceContext,
	endorsementStrings []string,
) (*ear.AttestationResult, error) {
	result := handler.CreateAttestationResult(SchemeName)
	evidence := ec.Evidence.AsMap()
	digestValue, ok := evidence["pcr-digest"]
	if !ok {
		err := handler.BadEvidence(
			"evidence does not contain %q entry",
			"pcr-digest",
		)
		return result, err
	}

	evidenceDigest, ok := digestValue.(string)
	if !ok {
		err := handler.BadEvidence(
			"wrong type value %q entry; expected string but found %T",
			"pcr-digest",
			digestValue,
		)
		return result, err
	}

	var endorsements Endorsements
	if err := endorsements.Populate(endorsementStrings); err != nil {
		return result, err
	}

	appraisal := result.Submods[SchemeName]
	appraisal.VeraisonAnnotatedEvidence = &evidence

	if endorsements.Digest == evidenceDigest {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		*appraisal.Status = ear.TrustTierAffirming
	}

	return result, nil
}

func synthKeysFromAttrs(scope string, tenantID string, attr json.RawMessage) ([]string, error) {
	var (
		nodeID string
		err    error
	)

	switch scope {
	case "software component":
		var att RefValAttr
		if err = json.Unmarshal(attr, &att); err != nil {
			return nil, fmt.Errorf("unable to extract sw component: %w", err)
		}
		nodeID = att.NodeID
	case "trust anchor":
		var att TaAttr
		if err = json.Unmarshal(attr, &att); err != nil {
			return nil, fmt.Errorf("unable to extract trust anchor: %w", err)
		}
		nodeID = att.NodeID
	default:
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}

	return []string{tpmEnactTrustLookupKey(tenantID, nodeID)}, nil
}

func parseKey(trustAnchor string) (*ecdsa.PublicKey, error) {
	var taEndorsement TrustAnchorEndorsement

	if err := json.Unmarshal([]byte(trustAnchor), &taEndorsement); err != nil {
		return nil, fmt.Errorf("could not decode trust anchor: %w", err)
	}

	buf, err := base64.StdEncoding.DecodeString(taEndorsement.Attr.Key)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %v", err)
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}

func tpmEnactTrustLookupKey(tenantID, nodeID string) string {
	absPath := []string{nodeID}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}
