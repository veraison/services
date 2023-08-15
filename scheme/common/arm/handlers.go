// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/veraison/ccatoken"
	parsec_cca "github.com/veraison/parsec/cca"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
)

type SwAttr struct {
	ImplID           []byte `cca:"CCA_SSD_PLATFORM.impl-id" psa:"PSA_IOT.impl-id" parcca:"PARSEC_CCA.impl-id"`
	Model            string `cca:"CCA_SSD_PLATFORM.hw-model" psa:"PSA_IOT.hw-model" parcca:"PARSEC_CCA.hw-model"`
	Vendor           string `cca:"CCA_SSD_PLATFORM.hw-vendor" psa:"PSA_IOT.hw-vendor" parcca:"PARSEC_CCA.hw-vendor"`
	MeasDesc         string `cca:"CCA_SSD_PLATFORM.measurement-desc" psa:"PSA_IOT.measurement-desc" parcca:"PARSEC_CCA.measurement-desc"`
	MeasurementType  string `cca:"CCA_SSD_PLATFORM.measurement-type" psa:"PSA_IOT.measurement-type" parcca:"PARSEC_CCA.measurement-type"`
	MeasurementValue []byte `cca:"CCA_SSD_PLATFORM.measurement-value" psa:"PSA_IOT.measurement-value" parcca:"PARSEC_CCA.measurement-value"`
	SignerID         []byte `cca:"CCA_SSD_PLATFORM.signer-id" psa:"PSA_IOT.signer-id" parcca:"PARSEC_CCA.signer-id"`
	Version          string `cca:"CCA_SSD_PLATFORM.version" psa:"PSA_IOT.version" parcca:"PARSEC_CCA.version"`
}

type TaAttr struct {
	Model    string `cca:"CCA_SSD_PLATFORM.hw-model" psa:"PSA_IOT.hw-model" parcca:"PARSEC_CCA.hw-model"`
	Vendor   string `cca:"CCA_SSD_PLATFORM.hw-vendor" psa:"PSA_IOT.hw-vendor" parcca:"PARSEC_CCA.hw-vendor"`
	VerifKey string `cca:"CCA_SSD_PLATFORM.iak-pub" psa:"PSA_IOT.iak-pub" parcca:"PARSEC_CCA.iak-pub"`
	ImplID   []byte `cca:"CCA_SSD_PLATFORM.impl-id" psa:"PSA_IOT.impl-id" parcca:"PARSEC_CCA.impl-id"`
	InstID   string `cca:"CCA_SSD_PLATFORM.inst-id" psa:"PSA_IOT.inst-id" parcca:"PARSEC_CCA.inst-id"`
}

type CcaPlatformCfg struct {
	ImplID []byte `cca:"CCA_SSD_PLATFORM.impl-id" parcca:"PARSEC_CCA.impl-id"`
	Model  string `cca:"CCA_SSD_PLATFORM.hw-model" parcca:"PARSEC_CCA.hw-model"`
	Vendor string `cca:"CCA_SSD_PLATFORM.hw-vendor" parcca:"PARSEC_CCA.hw-vendor"`
	Label  string `cca:"CCA_SSD_PLATFORM.platform-config-label" parcca:"PARSEC_CCA.platform-config-label"`
	Value  []byte `cca:"CCA_SSD_PLATFORM.platform-config-id" parcca:"PARSEC_CCA.platform-config-id"`
}

func SynthKeysFromRefValue(scheme string, tenantID string,
	refVal *handler.Endorsement,
) ([]string, error) {

	implID, err := common.GetImplID(scheme, refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	lookupKey := RefValLookupKey(scheme, tenantID, implID)
	log.Debugf("Scheme %s Plugin Reference Value Look Up Key= %s\n", scheme, lookupKey)

	return []string{lookupKey}, nil

}

func SynthKeysFromTrustAnchors(scheme string, tenantID string,
	ta *handler.Endorsement,
) ([]string, error) {

	implID, err := common.GetImplID(scheme, ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value: %w", err)
	}

	instID, err := common.GetInstID(scheme, ta.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	lookupKey := TaLookupKey(scheme, tenantID, implID, instID)
	log.Debugf("Scheme %s Plugin TA Look Up Key= %s\n", scheme, lookupKey)
	return []string{lookupKey}, nil
}

func GetTrustAnchorID(scheme string, token *proto.AttestationToken) (string, error) {
	var claims psatoken.IClaims

	switch scheme {
	case "PSA_IOT":
		var psaToken psatoken.Evidence

		err := psaToken.FromCOSE(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}
		claims = psaToken.Claims

	case "CCA_SSD_PLATFORM":
		var evidence ccatoken.Evidence

		err := evidence.FromCBOR(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}

		claims = evidence.PlatformClaims

	case "PARSEC_CCA":
		var evidence parsec_cca.Evidence

		err := evidence.FromCBOR(token.Data)
		if err != nil {
			return "", handler.BadEvidence(err)
		}
		claims = evidence.Pat.PlatformClaims
	default:
		return "", fmt.Errorf("invalid scheme argument to GetTrustAnchorID : %s", scheme)

	}

	return TaLookupKey(
		scheme,
		token.TenantId,
		MustImplIDString(claims),
		MustInstIDString(claims),
	), nil
}

func MatchSoftware(scheme string, evidence psatoken.IClaims, endorsements []handler.Endorsement) bool {
	var attr SwAttr

	var schemeJSON jsoniter.API

	switch scheme {
	case "PSA_IOT":
		schemeJSON = jsoniter.Config{TagKey: "psa"}.Froze()
	case "CCA_SSD_PLATFORM":
		schemeJSON = jsoniter.Config{TagKey: "cca"}.Froze()
	case "PARSEC_CCA":
		schemeJSON = jsoniter.Config{TagKey: "parcca"}.Froze()
	default:

	}
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

		if err := schemeJSON.Unmarshal(endorsement.Attributes, &attr); err != nil {
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

func FilterRefVal(endorsements []handler.Endorsement, key string) []handler.Endorsement {
	var refVal []handler.Endorsement
	for _, end := range endorsements {
		if end.SubType == key {
			refVal = append(refVal, end)
		}
	}
	return refVal
}

func GetPublicKeyFromTA(scheme string, trustAnchor string) (crypto.PublicKey, error) {
	var (
		endorsement handler.Endorsement
		ta          TaAttr
		schemeJSON  jsoniter.API
	)

	switch scheme {
	case "PSA_IOT":
		schemeJSON = jsoniter.Config{TagKey: "psa"}.Froze()
	case "CCA_SSD_PLATFORM":
		schemeJSON = jsoniter.Config{TagKey: "cca"}.Froze()
	case "PARSEC_CCA":
		schemeJSON = jsoniter.Config{TagKey: "parcca"}.Froze()
	default:
		return nil, fmt.Errorf("invalid scheme: %s", scheme)
	}
	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		return nil, fmt.Errorf("could not decode trust anchor: %w", err)
	}

	if err := schemeJSON.Unmarshal(endorsement.Attributes, &ta); err != nil {
		return nil, fmt.Errorf("could not unmarshal trust anchor: %w", err)
	}
	pem := ta.VerifKey

	pk, err := common.DecodePemSubjectPubKeyInfo([]byte(pem))
	if err != nil {
		return nil, fmt.Errorf("could not decode subject public key info: %w", err)
	}
	return pk, nil
}

func MatchPlatformConfig(scheme string, evidence psatoken.IClaims, endorsements []handler.Endorsement) bool {

	var (
		attr       CcaPlatformCfg
		schemeJSON jsoniter.API
	)

	switch scheme {
	case "CCA_SSD_PLATFORM":
		schemeJSON = jsoniter.Config{TagKey: "cca"}.Froze()
	case "PARSEC_CCA":
		schemeJSON = jsoniter.Config{TagKey: "parcca"}.Froze()
	default:
		log.Errorf("invalid scheme name: %s", scheme)
		return false
	}

	pfConfig, err := evidence.GetConfig()
	if err != nil {
		return false
	}
	if len(endorsements) > 1 {
		log.Errorf("got %d CCA configuration endorsements, want 1", len(endorsements))
		return false
	}

	if err := schemeJSON.Unmarshal(endorsements[0].Attributes, &attr); err != nil {
		log.Error("could not decode cca platform config in MatchPlatformConfig")
		return false
	}

	return bytes.Equal(pfConfig, attr.Value)
}
