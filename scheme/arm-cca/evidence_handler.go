// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
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
	ccaToken, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(token.Data)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	platformClaimsSet, err := common.ClaimsToMap(common.CcaPlatformWrapper{ccaToken.PlatformClaims}) // nolint:govet
	if err != nil {
		return nil, handler.BadEvidence(fmt.Errorf(
			"could not convert platform claims: %w", err))
	}

	realmClaimsSet, err := common.ClaimsToMap(common.CcaRealmWrapper{ccaToken.RealmClaims}) // nolint:govet
	if err != nil {
		return nil, handler.BadEvidence(fmt.Errorf(
			"could not convert realm claims: %w", err))
	}

	claims := map[string]interface{}{
		"platform": platformClaimsSet,
		"realm":    realmClaimsSet,
	}

	return claims, nil
}

// ValidateEvidenceIntegrity, decodes CCA collection and then invokes Verify API of ccatoken library
// which verifies the signature on the platform part of CCA collection, using supplied trust anchor
// and internally verifies the realm part of CCA token using realm public key extracted from
// realm token.
func (s EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	endorsementsStrings []string,
) error {
	ccaToken, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(token.Data)
	if err != nil {
		return handler.BadEvidence(err)
	}

	realmChallenge, err := ccaToken.RealmClaims.GetChallenge()
	if err != nil {
		return handler.BadEvidence(err)
	}

	// If the provided challenge was less than 64 bytes long, the RMM will
	// zero-pad pad it when generating the attestation token, so do the
	// same to the session nonce.
	sessionNonce := make([]byte, 64)
	copy(sessionNonce, token.Nonce)

	if !bytes.Equal(realmChallenge, sessionNonce) {
		return handler.BadEvidence(
			"freshness: realm challenge (%s) does not match session nonce (%s)",
			hex.EncodeToString(realmChallenge),
			hex.EncodeToString(token.Nonce),
		)
	}

	pk, err := arm.GetPublicKeyFromTA(SchemeName, trustAnchors[0])
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchor: %w", err)
	}

	if err = ccaToken.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}
	log.Debug("CCA platform token signature, realm token signature and cryptographic binding verified")
	return nil
}

func (s EvidenceHandler) AppraiseEvidence(
	ec *proto.EvidenceContext, endorsementsStrings []string,
) (*ear.AttestationResult, error) {
	endorsements, err := getEndorsementsFromString(endorsementsStrings)
	if err != nil {
		return nil, err
	}
	evidence := ec.Evidence.AsMap()
	result := handler.CreateAttestationResult("CCA_SSD_PLATFORM")

	/* Perform SubAttester Appraisal */
	claims, ok := evidence["platform"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to get platform claims: %w", handler.BadEvidence(err))
	}

	appraisal, err := platformAppraisal(claims, endorsements)
	if err != nil {
		return nil, err
	}
	result.Submods["CCA_SSD_PLATFORM"] = appraisal
	claims, ok = evidence["realm"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to get realm claims: %w", handler.BadEvidence(err))
	}

	appraisal, err = realmAppraisal(claims, endorsements)
	if err != nil {
		return nil, err
	}
	result.Submods["CCA_REALM"] = appraisal
	return result, nil
}

func getEndorsementsFromString(endorsementsStrings []string) ([]handler.Endorsement, error) {
	var endorsements []handler.Endorsement // nolint:prealloc
	for i, e := range endorsementsStrings {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}
		endorsements = append(endorsements, endorsement)
	}
	return endorsements, nil
}
