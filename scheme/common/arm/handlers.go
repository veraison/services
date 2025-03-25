// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/veraison/ccatoken/platform"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
)

type SwAttr struct {
	ImplID           []byte `json:"impl-id"`
	Model            string `json:"hw-model"`
	Vendor           string `json:"hw-vendor"`
	MeasDesc         string `json:"measurement-desc"`
	MeasurementType  string `json:"measurement-type"`
	MeasurementValue []byte `json:"measurement-value"`
	SignerID         []byte `json:"signer-id"`
	Version          string `json:"version"`
}

type TaAttr struct {
	Model    string `json:"hw-model"`
	Vendor   string `json:"hw-vendor"`
	VerifKey string `json:"iak-pub"`
	ImplID   []byte `json:"impl-id"`
	InstID   string `json:"inst-id"`
}

type CcaPlatformCfg struct {
	ImplID []byte `json:"impl-id"`
	Model  string `json:"hw-model"`
	Vendor string `json:"hw-vendor"`
	Label  string `json:"platform-config-label"`
	Value  []byte `json:"platform-config-id"`
}

func SynthKeysForPlatform(scheme string, tenantID string,
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

func GetPlatformReferenceIDs(
	scheme string,
	tenantID string,
	claims map[string]interface{},
) ([]string, error) {
	// Using the PSA specialisation here is ok because Implementation ID is
	// mandatory and shared by both PSA and CCA platform.
	platformClaims, err := common.MapToPSAClaims(claims)
	if err != nil {
		return nil, err
	}

	return []string{RefValLookupKey(
		scheme,
		tenantID,
		MustImplIDString(platformClaims),
	)}, nil
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

	verificationLookupKey := TaLookupKey(scheme, tenantID, implID, instID)
	log.Debugf("TA verification look up key: %s", verificationLookupKey)

	coservLookupKey := TaCoservLookupKey(scheme, tenantID, instID)
	log.Debugf("TA coserv look up key: %s", coservLookupKey)

	return []string{verificationLookupKey, coservLookupKey}, nil
}

func GetTrustAnchorID(scheme string, tenantID string, claims psatoken.IClaims) (string, error) {
	return TaLookupKey(
		scheme,
		tenantID,
		MustImplIDString(claims),
		MustInstIDString(claims),
	), nil
}

func MatchSoftware(scheme string, evidence psatoken.IClaims, endorsements []handler.Endorsement) bool {
	var attr SwAttr

	evidenceComponents := make(map[string]psatoken.ISwComponent)
	swComps, err := evidence.GetSoftwareComponents()
	if err != nil {
		return false
	}
	for _, c := range swComps {
		mval, err := c.GetMeasurementValue()
		if err != nil {
			return false
		}
		mtyp, err := c.GetMeasurementType()
		if err != nil {
			return false
		}
		key := base64.StdEncoding.EncodeToString(mval) + mtyp
		evidenceComponents[key] = c
	}
	matched := false
	for _, endorsement := range endorsements {
		// If we have Endorsements we assume they match to begin with
		matched = true

		if err := json.Unmarshal(endorsement.Attributes, &attr); err != nil {
			log.Error("could not decode sw attributes from endorsements")
			return false
		}

		key := base64.StdEncoding.EncodeToString(attr.MeasurementValue) + attr.MeasurementType
		evComp, ok := evidenceComponents[key]
		if !ok {
			matched = false
			break
		}

		evCompMeasurementType, _ := evComp.GetMeasurementType()
		evCompSignerID, _ := evComp.GetSignerID()
		evCompVersion, _ := evComp.GetVersion()

		log.Debugf("MeasurementType Evidence: %s, Endorsement: %s", evCompMeasurementType, attr.MeasurementType)
		typeMatched := attr.MeasurementType == "" || attr.MeasurementType == evCompMeasurementType
		sigMatched := attr.SignerID == nil || bytes.Equal(attr.SignerID, evCompSignerID)
		versionMatched := attr.Version == "" || attr.Version == evCompVersion

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
	)

	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		return nil, fmt.Errorf("for scheme, %s, could not decode trust anchor: %w", scheme, err)
	}

	if err := json.Unmarshal(endorsement.Attributes, &ta); err != nil {
		return nil, fmt.Errorf("could not unmarshal trust anchor: %w", err)
	}
	pem := ta.VerifKey

	pk, err := common.DecodePemSubjectPubKeyInfo([]byte(pem))
	if err != nil {
		return nil, fmt.Errorf("could not decode subject public key info: %w", err)
	}
	return pk, nil
}

func MatchPlatformConfig(scheme string, evidence platform.IClaims, endorsements []handler.Endorsement) bool {
	var attr CcaPlatformCfg
	pfConfig, err := evidence.GetConfig()
	if err != nil {
		return false
	}

	endorsementsLen := len(endorsements)

	switch endorsementsLen {
	case 0:
		log.Debugf("got no CCA configuration endorsement, accepting unconditionally")
		return true
	case 1:
		break
	default:
		log.Errorf("got %d CCA configuration endorsements, want 1", endorsementsLen)
		return false
	}

	if err := json.Unmarshal(endorsements[0].Attributes, &attr); err != nil {
		log.Error("could not decode cca platform config in MatchPlatformConfig")
		return false
	}

	return bytes.Equal(pfConfig, attr.Value)
}
