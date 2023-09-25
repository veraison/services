// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_ssd_platform

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
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

func (s EvidenceHandler) SynthKeysFromRefValue(
	tenantID string,
	refVal *handler.Endorsement,
) ([]string, error) {
	return arm.SynthKeysFromRefValue(SchemeName, tenantID, refVal)

}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	return arm.SynthKeysFromTrustAnchors(SchemeName, tenantID, ta)
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	return arm.GetTrustAnchorID(SchemeName, token)
}

func (s EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchor string,
) (*handler.ExtractedClaims, error) {

	var ccaToken ccatoken.Evidence

	if err := ccaToken.FromCBOR(token.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}

	var extracted handler.ExtractedClaims

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

	extracted.ClaimsSet = map[string]interface{}{
		"platform": platformClaimsSet,
		"realm":    realmClaimsSet,
	}

	extracted.ReferenceID = arm.RefValLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(ccaToken.PlatformClaims),
	)
	log.Debugf("extracted Reference ID Key = %s", extracted.ReferenceID)
	return &extracted, nil
}

// ValidateEvidenceIntegrity, decodes CCA collection and then invokes Verify API of ccatoken library
// which verifies the signature on the platform part of CCA collection, using supplied trust anchor
// and internally verifies the realm part of CCA token using realm public key extracted from
// realm token.
func (s EvidenceHandler) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchor string,
	endorsementsStrings []string,
) error {
	var (
		ccaToken ccatoken.Evidence
	)

	if err := ccaToken.FromCBOR(token.Data); err != nil {
		return handler.BadEvidence(err)
	}

	pk, err := arm.GetPublicKeyFromTA(SchemeName, trustAnchor)
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

	// TO DO: Handle Unprocessed evidence when new Attestation Result interface
	// is ready. Please see issue #105
	return result, err
}

func populateAttestationResult(
	result *ear.AttestationResult,
	evidence map[string]interface{},
	endorsements []handler.Endorsement,
) error {
	claims, err := common.MapToClaims(evidence["platform"].(map[string]interface{}))
	if err != nil {
		return err
	}

	appraisal := result.Submods[SchemeName]

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	appraisal.TrustVector.Hardware = ear.GenuineHardwareClaim

	rawLifeCycle, err := claims.GetSecurityLifeCycle()
	if err != nil {
		return handler.BadEvidence(err)
	}

	lifeCycle := psatoken.CcaLifeCycleToState(rawLifeCycle)
	if lifeCycle == psatoken.CcaStateSecured ||
		lifeCycle == psatoken.CcaStateNonCcaPlatformDebug {
		appraisal.TrustVector.InstanceIdentity = ear.TrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.ApprovedRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.HwKeysEncryptedSecretsClaim
	} else {
		appraisal.TrustVector.InstanceIdentity = ear.UntrustworthyInstanceClaim
		appraisal.TrustVector.RuntimeOpaque = ear.VisibleMemoryRuntimeClaim
		appraisal.TrustVector.StorageOpaque = ear.UnencryptedSecretsClaim
	}

	swComps := arm.FilterRefVal(endorsements, "CCA_SSD_PLATFORM.sw-component")
	match := arm.MatchSoftware(SchemeName, claims, swComps)
	if match {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Debug("matchSoftware Success")

	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Debug("matchSoftware Failed")
	}

	platformConfig := arm.FilterRefVal(endorsements, "CCA_SSD_PLATFORM.platform-config")
	match = arm.MatchPlatformConfig(SchemeName, claims, platformConfig)

	if match {
		appraisal.TrustVector.Configuration = ear.ApprovedConfigClaim
		log.Debug("matchPlatformConfig Success")

	} else {
		appraisal.TrustVector.Configuration = ear.UnsafeConfigClaim
		log.Debug("matchPlatformConfig Failed")
	}
	appraisal.UpdateStatusFromTrustVector()

	appraisal.VeraisonAnnotatedEvidence = &evidence

	return nil
}
