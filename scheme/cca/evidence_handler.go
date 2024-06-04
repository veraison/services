// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/veraison/ccatoken"
	ar "github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

type EvidenceHandler struct{}

func (s EvidenceHandler) GetName() string {
	return "cca-evidence-handler"
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

	var ccaToken ccatoken.Evidence

	if err := ccaToken.FromCBOR(token.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}

	platformClaimsSet, err := common.ClaimsToMap(ccaToken.PlatformClaims)
	if err != nil {
		return nil, handler.BadEvidence(fmt.Errorf(
			"could not convert platform claims: %w", err))
	}

	realmClaimsSet, err := common.ClaimsToMap(ccaToken.RealmClaims)
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
	var (
		ccaToken ccatoken.Evidence
	)

	if err := ccaToken.FromCBOR(token.Data); err != nil {
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
) (*ar.AttestationResult, error) {
	var endorsements []handler.Endorsement // nolint:prealloc
	var err error
	subSchemes := []string{"CCA_SSD_PLATFORM", "CCA_REALM"}
	result := handler.CreateAttestationResult(subSchemes[0])

	for _, subscheme := range subSchemes {
		endorsements, err = filterEndorsements(subscheme, endorsementsStrings)
		if err != nil {
			return nil, err
		}
		appraisal, err := createAppraisal(subscheme, result)
		if err != nil {
			return nil, err
		}

		subAttester, claims, err := getSubAttester(subscheme, ec)
		if err != nil {
			return nil, err
		}
		err = subAttester.PerformAppraisal(appraisal, claims, endorsements)
		if err != nil {
			return nil, err
		}
	}

	return result, err
}

func filterEndorsements(subscheme string, endorsementsStrings []string) ([]handler.Endorsement, error) {
	var endorsements []handler.Endorsement
	for i, e := range endorsementsStrings {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}
		if endorsement.SubScheme == subscheme {
			endorsements = append(endorsements, endorsement)
		}
	}
	return endorsements, nil
}

func getSubAttester(
	subscheme string,
	ec *proto.EvidenceContext,
) (ISubAttester, map[string]interface{}, error) {
	var claims map[string]interface{}
	evidence := ec.Evidence.AsMap()
	switch subscheme {
	case "CCA_SSD_PLATFORM":
		claims = evidence["platform"].(map[string]interface{})
		return &Cca_platform_attester{}, claims, nil
	case "CCA_REALM":
		claims = evidence["realm"].(map[string]interface{})
		return &Cca_realm_attester{}, claims, nil
	default:
		return nil, claims, fmt.Errorf("invalid scheme: %s", subscheme)
	}
}

func createAppraisal(submodname string, ear *ar.AttestationResult) (*ar.Appraisal, error) {
	submod, ok := ear.Submods[submodname]
	if !ok {
		var status ar.TrustTier
		submod = &ar.Appraisal{
			TrustVector: &ar.TrustVector{},
			Status:      &status,
		}
		ear.Submods[submodname] = submod
	}
	return submod, nil
}
