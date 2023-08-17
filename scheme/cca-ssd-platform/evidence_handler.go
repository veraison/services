// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_ssd_platform

import (
	"bytes"
	"encoding/base64"
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

type SwAttr struct {
	ImplID           []byte `json:"CCA_SSD_PLATFORM.impl-id"`
	Model            string `json:"CCA_SSD_PLATFORM.hw-model"`
	Vendor           string `json:"CCA_SSD_PLATFORM.hw-vendor"`
	MeasDesc         string `json:"CCA_SSD_PLATFORM.measurement-desc"`
	MeasurementType  string `json:"CCA_SSD_PLATFORM.measurement-type"`
	MeasurementValue []byte `json:"CCA_SSD_PLATFORM.measurement-value"`
	SignerID         []byte `json:"CCA_SSD_PLATFORM.signer-id"`
	Version          string `json:"CCA_SSD_PLATFORM.version"`
}

type CcaPlatformCfg struct {
	ImplID []byte `json:"CCA_SSD_PLATFORM.impl-id"`
	Model  string `json:"CCA_SSD_PLATFORM.hw-model"`
	Vendor string `json:"CCA_SSD_PLATFORM.hw-vendor"`
	Label  string `json:"CCA_SSD_PLATFORM.platform-config-label"`
	Value  []byte `json:"CCA_SSD_PLATFORM.platform-config-id"`
}

type Endorsements struct {
	Scheme  string          `json:"scheme"`
	Type    string          `json:"type"`
	SubType string          `json:"subType"`
	Attr    json.RawMessage `json:"attributes"`
}

type TaAttr struct {
	Model    string `json:"CCA_SSD_PLATFORM.hw-model"`
	Vendor   string `json:"CCA_SSD_PLATFORM.hw-vendor"`
	VerifKey string `json:"CCA_SSD_PLATFORM.iak-pub"`
	ImplID   []byte `json:"CCA_SSD_PLATFORM.impl-id"`
	InstID   string `json:"CCA_SSD_PLATFORM.inst-id"`
}

type TaEndorsements struct {
	Scheme  string `json:"scheme"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
	Attr    TaAttr `json:"attributes"`
}

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

	implID, err := common.GetImplID("CCA_SSD_PLATFORM", refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	lookupKey := arm.RefValLookupKey(SchemeName, tenantID, implID)
	log.Debug("CCA Plugin CCA Reference Value Look Up Key= %s\n", lookupKey)

	return []string{lookupKey}, nil
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	implID, err := common.GetImplID("CCA_SSD_PLATFORM", ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	instID, err := common.GetInstID("CCA_SSD_PLATFORM", ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	lookupKey := arm.TaLookupKey(SchemeName, tenantID, implID, instID)
	log.Debug("CCA Plugin TA CCA Look Up Key= %s\n", lookupKey)
	return []string{lookupKey}, nil
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var ccaToken ccatoken.Evidence

	err := ccaToken.FromCBOR(token.Data)
	if err != nil {
		return "", handler.BadEvidence(err)
	}

	return arm.TaLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(ccaToken.PlatformClaims),
		arm.MustInstIDString(ccaToken.PlatformClaims),
	), nil
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

	claimsSet, err := common.ClaimsToMap(ccaToken.PlatformClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	extracted.ClaimsSet = claimsSet

	extracted.ReferenceID = arm.RefValLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(ccaToken.PlatformClaims),
	)
	log.Debug("extracted Reference ID Key = %s", extracted.ReferenceID)
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
		endorsement TaEndorsements
		ccaToken    ccatoken.Evidence
	)

	if err := ccaToken.FromCBOR(token.Data); err != nil {
		return handler.BadEvidence(err)
	}

	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		return fmt.Errorf("could not decode trust anchor: %w", err)
	}

	ta := endorsement.Attr.VerifKey
	pk, err := common.DecodePemSubjectPubKeyInfo([]byte(ta))
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
	var endorsements []Endorsements // nolint:prealloc

	result := handler.CreateAttestationResult(SchemeName)

	for i, e := range endorsementsStrings {
		var endorsement Endorsements

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

func mapToClaims(in map[string]interface{}) (psatoken.IClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return psatoken.DecodeJSONClaims(data)
}

func populateAttestationResult(
	result *ear.AttestationResult,
	evidence map[string]interface{},
	endorsements []Endorsements,
) error {
	claims, err := mapToClaims(evidence)
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

	swComps := filterRefVal(endorsements, "CCA_SSD_PLATFORM.sw-component")
	match := matchSoftware(claims, swComps)
	if match {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Debug("matchSoftware Success")

	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Debug("matchSoftware Failed")
	}

	platformConfig := filterRefVal(endorsements, "CCA_SSD_PLATFORM.platform-config")
	match = matchPlatformConfig(claims, platformConfig)

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

func filterRefVal(endorsements []Endorsements, key string) []Endorsements {
	var refVal []Endorsements
	for _, end := range endorsements {
		if end.SubType == key {
			refVal = append(refVal, end)
		}
	}
	return refVal
}

func matchSoftware(evidence psatoken.IClaims, endorsements []Endorsements) bool {
	evidenceComponents := make(map[string]psatoken.SwComponent)

	swComps, err := evidence.GetSoftwareComponents()
	if err != nil {
		return false
	}

	for _, c := range swComps {
		key := base64.StdEncoding.EncodeToString(*c.MeasurementValue) + (*c.MeasurementType)
		evidenceComponents[key] = c
	}
	matched := false
	for _, endorsement := range endorsements {
		// If we have Endorsements we assume they match to begin with
		matched = true
		var attr SwAttr
		if err := json.Unmarshal(endorsement.Attr, &attr); err != nil {
			log.Error("could not decode sw attributes from endorsements")
			return false
		}

		key := base64.StdEncoding.EncodeToString(attr.MeasurementValue) + attr.MeasurementType
		evComp, ok := evidenceComponents[key]
		if !ok {
			matched = false
			break
		}

		log.Debug("MeasurementType Evidence: %s, Endorsement: %s", *evComp.MeasurementType, attr.MeasurementType)
		typeMatched := attr.MeasurementType == "" || attr.MeasurementType == *evComp.MeasurementType
		sigMatched := attr.SignerID == nil || bytes.Equal(attr.SignerID, *evComp.SignerID)
		versionMatched := attr.Version == "" || attr.Version == *evComp.Version

		if !(typeMatched && sigMatched && versionMatched) {
			matched = false
			break
		}
	}
	return matched
}

func matchPlatformConfig(evidence psatoken.IClaims, endorsements []Endorsements) bool {

	pfConfig, err := evidence.GetConfig()
	if err != nil {
		return false
	}
	if len(endorsements) > 1 {
		log.Error("got %d CCA configuration endorsements, want 1", len(endorsements))
		return false
	}
	var attr CcaPlatformCfg
	if err := json.Unmarshal(endorsements[0].Attr, &attr); err != nil {
		log.Error("could not decode cca platform config in matchPlatformConfig")
		return false
	}

	return bytes.Equal(pfConfig, attr.Value)
}
