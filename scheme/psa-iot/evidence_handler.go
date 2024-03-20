// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

type EvidenceHandler struct{}

func (s EvidenceHandler) GetName() string {
	return "psa-evidence-handler"
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
) (*handler.ExtractedClaims, error) {
	var psaToken psatoken.Evidence

	if err := psaToken.FromCOSE(token.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}

	var extracted handler.ExtractedClaims

	claimsSet, err := common.ClaimsToMap(psaToken.Claims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	extracted.ClaimsSet = claimsSet

	extracted.ReferenceIDs = []string{arm.RefValLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(psaToken.Claims),
	)}
	log.Printf("\n Extracted SW ID Key = %s", extracted.ReferenceIDs)
	return &extracted, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchors []string,
	endorsementsStrings []string,
) error {
	var (
		psaToken psatoken.Evidence
	)

	if err := psaToken.FromCOSE(token.Data); err != nil {
		return handler.BadEvidence(err)
	}

	psaNonce, err := psaToken.Claims.GetNonce()
	if err != nil {
		return handler.BadEvidence(err)
	}
	if !bytes.Equal(psaNonce, token.Nonce) {
		return handler.BadEvidence(
			"freshness: psa-nonce (%s) does not match session nonce (%s)",
			hex.EncodeToString(psaNonce),
			hex.EncodeToString(token.Nonce),
		)
	}

	pk, err := arm.GetPublicKeyFromTA(SchemeName, trustAnchors[0])
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchor: %w", err)
	}

	if err = psaToken.Verify(pk); err != nil {
		return handler.BadEvidence(err)
	}
	log.Println("\n Token Signature Verified")
	return nil
}

func (s EvidenceHandler) AppraiseEvidence(
	ec *proto.EvidenceContext, endorsementsStrings []string,
) (*ear.AttestationResult, error) {
	var endorsements []handler.Endorsement // nolint:prealloc

	result := handler.CreateAttestationResult(SchemeName)

	for i, e := range endorsementsStrings {
		var endorsement handler.Endorsement

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		endorsements = append(endorsements, endorsement)
	}

	err := populateAttestationResult(result, ec.Evidence.AsMap(), endorsements)

	return result, err
}

func populateAttestationResult(
	result *ear.AttestationResult,
	evidence map[string]interface{},
	endorsements []handler.Endorsement,
) error {
	claims, err := common.MapToClaims(evidence)
	if err != nil {
		return handler.BadEvidence(err)
	}

	appraisal := result.Submods[SchemeName]

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	rawLifeCycle, err := claims.GetSecurityLifeCycle()
	if err != nil {
		return handler.BadEvidence(err)
	}

	lifeCycle := psatoken.PsaLifeCycleToState(rawLifeCycle)
	if lifeCycle == psatoken.PsaStateSecured || lifeCycle == psatoken.PsaStateNonPsaRotDebug {
		appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.ApprovedRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.HwKeysEncryptedSecretsClaim
	} else {
		appraisal.TrustVector.InstanceIdentity = ear.UntrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.UnencryptedSecretsClaim
	}

	match := arm.MatchSoftware(SchemeName, claims, endorsements)
	if match {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Println("\n matchSoftware Success")

	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Println("\n matchSoftware Failed")
	}

	appraisal.UpdateStatusFromTrustVector()

	appraisal.VeraisonAnnotatedEvidence = &evidence

	return nil
}
