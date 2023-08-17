// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/ear"
	"github.com/veraison/go-cose"
	parsec_cca "github.com/veraison/parsec/cca"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

const (
	ScopeTrustAnchor = "trust anchor"
	ScopeRefValues   = "ref values"
)

type SwAttr struct {
	ImplID           []byte `json:"PARSEC_CCA.impl-id"`
	Model            string `json:"PARSEC_CCA.hw-model"`
	Vendor           string `json:"PARSEC_CCA.hw-vendor"`
	MeasDesc         string `json:"PARSEC_CCA.measurement-desc"`
	MeasurementType  string `json:"PARSEC_CCA.measurement-type"`
	MeasurementValue []byte `json:"PARSEC_CCA.measurement-value"`
	SignerID         []byte `json:"PARSEC_CCA.signer-id"`
	Version          string `json:"PARSEC_CCA.version"`
}

type CcaPlatformCfg struct {
	ImplID []byte `json:"PARSEC_CCA.impl-id"`
	Model  string `json:"PARSEC_CCA.hw-model"`
	Vendor string `json:"PARSEC_CCA.hw-vendor"`
	Label  string `json:"PARSEC_CCA.platform-config-label"`
	Value  []byte `json:"PARSEC_CCA.platform-config-id"`
}

type Endorsements struct {
	Scheme  string          `json:"scheme"`
	Type    string          `json:"type"`
	SubType string          `json:"subType"`
	Attr    json.RawMessage `json:"attributes"`
}

type TaAttr struct {
	Model    string `json:"PARSEC_CCA.hw-model"`
	Vendor   string `json:"PARSEC_CCA.hw-vendor"`
	VerifKey string `json:"PARSEC_CCA.iak-pub"`
	ImplID   []byte `json:"PARSEC_CCA.impl-id"`
	InstID   string `json:"PARSEC_CCA.inst-id"`
}

type TaEndorsements struct {
	Scheme  string          `json:"scheme"`
	Type    string          `json:"type"`
	SubType string          `json:"sub_type"`
	Attr    json.RawMessage `json:"attributes"`
}

type EvidenceHandler struct{}

func (s EvidenceHandler) GetName() string {
	return "parsec-cca-evidence-handler"
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

	implID, err := common.GetImplID("PARSEC_CCA", refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	lookupKey := arm.RefValLookupKey(SchemeName, tenantID, implID)
	log.Debugf("PARSEC CCA Plugin Reference Value Look Up Key= %s\n", lookupKey)

	return []string{lookupKey}, nil
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	implID, err := common.GetImplID("PARSEC_CCA", ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	instID, err := common.GetInstID("PARSEC_CCA", ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	lookupKey := arm.TaLookupKey(SchemeName, tenantID, implID, instID)
	log.Debugf("PARSEC CCA Plugin TA Look Up Key= %s\n", lookupKey)
	return []string{lookupKey}, nil
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var evidence parsec_cca.Evidence

	err := evidence.FromCBOR(token.Data)
	if err != nil {
		return "", handler.BadEvidence(err)
	}

	return arm.TaLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(evidence.Pat.PlatformClaims),
		arm.MustInstIDString(evidence.Pat.PlatformClaims),
	), nil
}

func (s EvidenceHandler) ExtractClaims(token *proto.AttestationToken, trustAnchor string) (*handler.ExtractedClaims, error) {
	var (
		extracted handler.ExtractedClaims
		evidence  parsec_cca.Evidence
		claimsSet = make(map[string]interface{})
		kat       = make(map[string]interface{})
	)

	if err := evidence.FromCBOR(token.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}
	kat["nonce"] = *evidence.Kat.Nonce
	key := evidence.Kat.Cnf.COSEKey
	ck, err := key.MarshalCBOR()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}

	kat["akpub"] = base64.StdEncoding.EncodeToString(ck)

	claimsSet["kat"] = kat
	pmap, err := common.ClaimsToMap(evidence.Pat.PlatformClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	claimsSet["cca.platform"] = pmap
	rmap, err := common.ClaimsToMap(evidence.Pat.RealmClaims)
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	claimsSet["cca.realm"] = rmap

	extracted.ClaimsSet = claimsSet

	extracted.ReferenceID = arm.RefValLookupKey(
		SchemeName,
		token.TenantId,
		arm.MustImplIDString(evidence.Pat.PlatformClaims),
	)
	log.Debugf("extracted Reference ID Key = %s", extracted.ReferenceID)
	return &extracted, nil
}

func (s EvidenceHandler) ValidateEvidenceIntegrity(token *proto.AttestationToken, trustAnchor string, endorsements []string) error {
	var (
		endorsement TaEndorsements
		evidence    parsec_cca.Evidence
	)

	if err := evidence.FromCBOR(token.Data); err != nil {
		return handler.BadEvidence(err)
	}

	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		return fmt.Errorf("could not decode trust anchor: %w", err)
	}
	var ta TaAttr
	if err := json.Unmarshal(endorsement.Attr, &ta); err != nil {
		return fmt.Errorf("could not unmarshal cca trust anchor: %w", err)
	}
	pem := ta.VerifKey
	pk, err := common.DecodePemSubjectPubKeyInfo([]byte(pem))
	if err != nil {
		return fmt.Errorf("could not get public key from trust anchor: %w", err)
	}

	if err = evidence.Verify(pk); err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	log.Debug("Parsec CCA token signature, verified")
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
	return result, err
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
	kmap, ok := evidence["kat"]
	if !ok {
		return handler.BadEvidence(errors.New("no key attestation map in the evidence"))
	}
	kat := kmap.(map[string]interface{})

	key, ok := kat["akpub"]
	if !ok {
		return handler.BadEvidence(errors.New("no key in the evidence"))
	}
	var COSEKey cose.Key

	kb, err := base64.StdEncoding.DecodeString(key.(string))
	if err != nil {
		return handler.BadEvidence(err)
	}
	err = COSEKey.UnmarshalCBOR(kb)
	if err != nil {
		return handler.BadEvidence(err)
	}
	// Extract Public Key and set the Veraison Extension
	pk, err := COSEKey.PublicKey()
	if err != nil {
		return handler.BadEvidence(err)
	}

	if err := appraisal.SetKeyAttestation(pk); err != nil {
		return fmt.Errorf("setting extracted public key: %w", err)
	}

	cp, ok := evidence["cca.platform"]
	if !ok {
		return handler.BadEvidence(errors.New("no cca platform in the evidence"))
	}
	pmap := cp.(map[string]interface{})
	claims, err := mapToClaims(pmap)
	if err != nil {
		return handler.BadEvidence(err)
	}

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

	swComps := filterRefVal(endorsements, "PARSEC_CCA.sw-component")
	match := matchSoftware(claims, swComps)
	if match {
		appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Debug("matchSoftware Success")

	} else {
		appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Debug("matchSoftware Failed")
	}

	platformConfig := filterRefVal(endorsements, "PARSEC_CCA.platform-config")
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

func mapToClaims(in map[string]interface{}) (psatoken.IClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return psatoken.DecodeJSONClaims(data)
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

		log.Debugf("MeasurementType Evidence: %s, Endorsement: %s", *evComp.MeasurementType, attr.MeasurementType)
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
