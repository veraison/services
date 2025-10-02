// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"

	tpm2 "github.com/google/go-tpm/tpm2"

	"github.com/veraison/ear"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
)

type EvidenceHandler struct{}

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

	claims := make(map[string]interface{})
	claims["pcr-selection"] = pcrs
	claims["hash-algorithm"] = int64(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.Hash)
	claims["firmware-version"] = decoded.AttestationData.FirmwareVersion
	claims["node-id"] = decoded.NodeId.String()
	claims["pcr-digest"] = []byte(decoded.AttestationData.AttestedQuoteInfo.PCRDigest)

	return claims, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	endorsements []string,
) error {
	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return handler.BadEvidence("could not decode token: %w", err)
	}

	pubKey, err := parseKey(trustAnchors[0])
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
	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		*appraisal.Status = ear.TrustTierContraindicated
	}

	return result, nil
}

func parseKey(trustAnchor string) (*ecdsa.PublicKey, error) {
	var taEndorsement TrustAnchorEndorsement

	if err := json.Unmarshal([]byte(trustAnchor), &taEndorsement); err != nil {
		return nil, fmt.Errorf("could not decode trust anchor: %w", err)
	}

	key, err := common.DecodePemSubjectPubKeyInfo([]byte(taEndorsement.Attr.Key))
	if err != nil {
		return nil, err
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}
