// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_realm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/ccatoken"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
)

type EvidenceHandler struct{}

type RealmAttr struct {
	Vendor           string   `json:"cca-realm.vendor"`
	Model            string   `json:"cca-realm.model"`
	RealmID          string   `json:"cca-realm.id"`
	AlgID            string   `json:"cca-realm.alg-id"`
	MeasurementArray [][]byte `json:"cca-realm.measurement-array"`
}

func (s EvidenceHandler) GetName() string {
	return "cca-realm-evidence-handler"
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
	var realm RealmAttr

	attr := refVal.Attributes
	err := json.Unmarshal(attr, &realm)
	if err != nil {

		return nil, fmt.Errorf("unable to UnMarshal Realm Attributes %w", err)
	}
	if realm.MeasurementArray == nil {
		return nil, fmt.Errorf("no measurements in Realm Endorsements %w", err)
	}

	rim := base64.StdEncoding.EncodeToString(realm.MeasurementArray[0])
	log.Debugf("base64 encoded rim value = %s", rim)

	lookupKey := RefValLookupKey(SchemeName, tenantID, rim)
	log.Debugf("Scheme %s Plugin Reference Value Look Up Key= %s\n", SchemeName, lookupKey)

	return []string{lookupKey}, nil
}

func RefValLookupKey(schemeName, tenantID, rim string) string {
	absPath := []string{rim}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func (s EvidenceHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	return nil, fmt.Errorf("unsupported method SynthKeysFromTrustAnchor() for realm verification plugin")
}

func (s EvidenceHandler) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	log.Debug("Yogesh: REALM: GetTrustAnchorID invoked")
	return "", nil
}

// ExtractClaims extract Realm Claims and set the extracted Reference ID
func (s EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchor string,
) (*handler.ExtractedClaims, error) {

	var ccaToken ccatoken.Evidence
	log.Debug("Yogesh: REALM: ExtractClaims invoked")
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

	brim, err := ccaToken.RealmClaims.GetInitialMeasurement()
	if err != nil {
		return nil, handler.BadEvidence(err)
	}
	rim := base64.StdEncoding.EncodeToString(brim)
	log.Debugf("base64 encoded rim value = %s", rim)

	extracted.ReferenceID = RefValLookupKey(
		SchemeName,
		token.TenantId,
		rim)
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
	log.Debug(" Yogesh: Realm: ValidateEvidence Integrity invoked")
	// Please note this part of Evidence Validation is already done in the Platform part
	log.Debug("CCA platform token signature, realm token signature and cryptographic binding verified")
	return nil
}

func (s EvidenceHandler) AppraiseEvidence(
	ec *proto.EvidenceContext, endorsementsStrings []string,
) (*ear.AttestationResult, error) {
	var endorsements []handler.Endorsement // nolint:prealloc
	log.Debug(" Yogesh: Realm: AppraiseEvidence  invoked")
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
	claims, err := mapToClaims(evidence["realm"].(map[string]interface{}))
	if err != nil {
		return err
	}

	appraisal := result.Submods[SchemeName]

	// Match RIM Values first
	brim, err := claims.GetInitialMeasurement()
	if err != nil {
		return fmt.Errorf("failed to get Realm Initial Measurements from Claims")
	}
	rim := base64.StdEncoding.EncodeToString(brim)
	log.Debugf("base64 encoded rim value = %s", rim)

	for _, endorsement := range endorsements {
		var realmEnd RealmAttr
		json.Unmarshal(endorsement.Attributes, &realmEnd)
		er := realmEnd.MeasurementArray[0]
		erim := base64.StdEncoding.EncodeToString(er)
		log.Debugf("base64 encoded endorsement rim value = %s", erim)

		if rim == erim {
			appraisal.TrustVector.Executables = ear.ApprovedBootClaim
			log.Debug("realm Initial Measurement match Success")

		} else {
			appraisal.TrustVector.Executables = ear.ContraindicatedRuntimeClaim
			log.Debug("realm Initial Measurement match Failed")
		}

		if appraisal.TrustVector.Executables == ear.ApprovedBootClaim {
			execmatch, err := matchExecutables(claims, realmEnd)
			if err != nil {
				appraisal.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
				break
			}
			// Match REM Values to express Execution
			if execmatch {
				appraisal.TrustVector.Executables = ear.ApprovedRuntimeClaim
				log.Debug("Boot claim and run time both succeedded")

			} else {
				appraisal.TrustVector.Executables = ear.UnsafeRuntimeClaim
				log.Debug("Boot claim succeedded but run time both  Failed")
			}
			break
		}
	}
	appraisal.UpdateStatusFromTrustVector()

	return nil
}

func mapToClaims(in map[string]interface{}) (*ccatoken.RealmClaims, error) {
	var cca ccatoken.RealmClaims
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	err = cca.FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("unable to map claims for RealmClaims %w", err)
	}
	return &cca, nil
}

func matchExecutables(rc *ccatoken.RealmClaims, re RealmAttr) (bool, error) {

	if rc == nil {
		return false, fmt.Errorf("no realm claims in matchExecutables")
	}
	rems, err := rc.GetExtensibleMeasurements()
	if err != nil {
		return false, fmt.Errorf("unable to matchExecutables %w", err)
	}

	matched := false
	for _, meas := range rems {
		matched = false
		for _, emeas := range re.MeasurementArray {
			if bytes.Equal(meas, emeas) {
				matched = true
			}
		}
		if !matched {
			return !matched, nil
		}
	}

	return matched, nil
}
