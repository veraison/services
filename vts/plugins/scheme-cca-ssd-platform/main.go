// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/ccatoken"
	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"github.com/veraison/services/vts/plugins/common"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type SwAttr struct {
	ImplID    []byte `json:"cca.impl-id"`
	Model     string `json:"cca.hw-model"`
	Vendor    string `json:"cca.hw-vendor"`
	MeasDesc  uint64 `json:"cca.measurement-desc"`
	MeasType  string `json:"cca.measurement-type"`
	MeasValue []byte `json:"cca.measurement-value"`
	SignerID  []byte `json:"cca.signer-id"`
	Version   string `json:"cca.version"`
}

type CcaPlatformCfg struct {
	ImplID []byte `json:"cca.impl-id"`
	Model  string `json:"cca.hw-model"`
	Vendor string `json:"cca.hw-vendor"`
	Label  string `json:"cca.platform-config-label"`
	Value  []byte `json:"cca.platform-config-id"`
}

type Endorsements struct {
	Scheme  string          `json:"scheme"`
	Type    string          `json:"type"`
	SubType string          `json:"sub_type"`
	Attr    json.RawMessage `json:"attributes"`
}

type TaAttr struct {
	Model    string `json:"cca.hw-model"`
	Vendor   string `json:"cca.hw-vendor"`
	VerifKey string `json:"cca.iak-pub"`
	ImplID   []byte `json:"cca.impl-id"`
	InstID   string `json:"cca.inst-id"`
}

type TaEndorsements struct {
	Scheme  string `json:"scheme"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
	Attr    TaAttr `json:"attributes"`
}

const SchemeName = "CCA_SSD_PLATFORM"

type Scheme struct{}

func (s Scheme) GetName() string {
	return SchemeName
}

func (s Scheme) SynthKeysFromRefValue(
	tenantID string,
	refVal *proto.Endorsement,
) ([]string, error) {
	var (
		implID string
		fields map[string]*structpb.Value
		err    error
	)

	fields, err = common.GetFieldsFromParts(refVal.GetAttributes())
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value abs-path: %w", err)
	}

	implID, err = common.GetMandatoryPathSegment("cca.impl-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value abs-path: %w", err)
	}

	finalstr := ccaReferenceLookupKey(tenantID, implID)
	log.Printf("CCA Plugin CCA Reference Value Look Up Key= %s\n", finalstr)

	return []string{ccaReferenceLookupKey(tenantID, implID)}, nil
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
	var (
		instID string
		implID string
		fields map[string]*structpb.Value
		err    error
	)

	fields, err = common.GetFieldsFromParts(ta.GetAttributes())
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	implID, err = common.GetMandatoryPathSegment("cca.impl-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	instID, err = common.GetMandatoryPathSegment("cca.inst-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	finalstr := ccaTaLookupKey(tenantID, implID, instID)
	log.Printf("CCA Plugin TA CCA Look Up Key= %s\n", finalstr)
	return []string{ccaTaLookupKey(tenantID, implID, instID)}, nil
}

func (s Scheme) GetSupportedMediaTypes() []string {
	return []string{
		"application/eat-collection; profile=http://arm.com/CCA-SSD/1.0.0",
	}
}

func (s Scheme) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var ccaToken ccatoken.Evidence

	err := ccaToken.FromCBOR(token.Data)
	if err != nil {
		return "", err
	}

	return ccaTaLookupKey(
		token.TenantId,
		MustImplIDString(ccaToken.PlatformClaims),
		MustInstIDString(ccaToken.PlatformClaims),
	), nil
}

func (s Scheme) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchor string,
) (*scheme.ExtractedClaims, error) {

	var ccaToken ccatoken.Evidence

	if err := ccaToken.FromCBOR(token.Data); err != nil {
		return nil, err
	}

	var extracted scheme.ExtractedClaims

	claimsSet, err := claimsToMap(ccaToken.PlatformClaims)
	if err != nil {
		return nil, err
	}

	extracted.ClaimsSet = claimsSet

	unprocessedclaimsSet, err := claimsToMap(ccaToken.RealmClaims)
	if err != nil {
		return nil, err
	}
	extracted.UnprocessedClaimsSet = unprocessedclaimsSet

	extracted.ReferenceID = ccaReferenceLookupKey(
		token.TenantId,
		MustImplIDString(ccaToken.PlatformClaims),
	)
	log.Printf("\n Extracted Reference ID Key = %s", extracted.ReferenceID)
	return &extracted, nil
}

// ValidateEvidenceIntegrity, decodes CCA collection and then invokes Verify API of ccatoken library.
// which verifies the signature on the platform part of CCA collection, using supplied trust anchor
// and internally verifies the realm part of CCA token using realm public key extracted from
// realm token.
func (s Scheme) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchor string,
	endorsementsStrings []string,
) error {
	var endorsement TaEndorsements

	if err := json.Unmarshal([]byte(trustAnchor), &endorsement); err != nil {
		log.Println("Could not decode Endorsements in ExtractVerifiedClaims")
		return fmt.Errorf("could not decode endorsement: %w", err)
	}
	ta := endorsement.Attr.VerifKey
	block, rest := pem.Decode([]byte(ta))

	if block == nil {
		log.Println("could not get TA PEM Block during validating evidence integrity")
		return errors.New("could not extract trust anchor PEM block")
	}

	if len(rest) != 0 {
		return errors.New("trailing data found after PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return fmt.Errorf("unsupported key type %q", block.Type)
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	var ccaToken ccatoken.Evidence

	if err = ccaToken.FromCBOR(token.Data); err != nil {
		return err
	}

	if err = ccaToken.Verify(pk); err != nil {
		return err
	}
	log.Println("\n CCA platform token signature, realm token signature and cryptographic binding verified")
	return nil
}

func (s Scheme) AppraiseEvidence(
	ec *proto.EvidenceContext, endorsementsStrings []string,
) (*ear.AttestationResult, error) {
	var endorsements []Endorsements

	result := ear.NewAttestationResult()

	for i, e := range endorsementsStrings {
		var endorsement Endorsements

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		endorsements = append(endorsements, endorsement)
	}

	err := populateAttestationResult(result, ec.Evidence.AsMap(), endorsements)

	// TO DO: Need to populate Unprocessed Evidence in a suitable format in AR
	/*
		// Unprocessed evidence should be a JSON Byte Array, which can be UnMarshalled by EAR Library
		result.unprocessed_evidence = json.Marshal(ec.UpEvidence.AsMap())
	*/

	return result, err
}

type ClaimMapper interface {
	ToJSON() ([]byte, error)
}

func claimsToMap(mapper ClaimMapper) (map[string]interface{}, error) {
	data, err := mapper.ToJSON()
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	return out, err
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

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	result.TrustVector.Hardware = ear.GenuineHardwareClaim

	swComps := filterRefVal(endorsements, "cca.sw-component")
	match := matchSoftware(claims, swComps)
	if match {
		result.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Println("\n matchSoftware Success")

	} else {
		result.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Println("\n matchSoftware Failed")
	}

	platformConfig := filterRefVal(endorsements, "cca.platform-config")
	match = matchPlatformConfig(claims, platformConfig)

	if match {
		result.TrustVector.Configuration = ear.ApprovedConfigClaim
		log.Println("\n matchPlatformConfig Success")

	} else {
		result.TrustVector.Configuration = ear.UnsafeConfigClaim
		log.Println("\n matchPlatformConfig Failed")
	}
	result.UpdateStatusFromTrustVector()

	result.VeraisonProcessedEvidence = &evidence

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
		key := base64.StdEncoding.EncodeToString(*c.MeasurementValue)
		evidenceComponents[key] = c
	}
	matched := false
	for _, endorsement := range endorsements {
		// If we have Endorsements we assume they match to begin with
		matched = true
		var attr SwAttr
		if err := json.Unmarshal(endorsement.Attr, &attr); err != nil {
			log.Println("Could not decode sw attributes from endorsements")
			return false
		}

		key := base64.StdEncoding.EncodeToString(attr.MeasValue)
		evComp, ok := evidenceComponents[key]
		if !ok {
			matched = false
			break
		}

		log.Printf("MeasType Evidence: %s, Endorsement: %s", *evComp.MeasurementType, attr.MeasType)
		typeMatched := attr.MeasType == "" || attr.MeasType == *evComp.MeasurementType
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
		log.Printf("matchPlatformConfig failed number of cca config %d > 1 ", len(endorsements))
	}
	var attr CcaPlatformCfg
	if err := json.Unmarshal(endorsements[0].Attr, &attr); err != nil {
		log.Println("Could not decode cca platform config in matchPlatformConfig")
		return false
	}

	return bytes.Equal(pfConfig, attr.Value)
}

func ccaReferenceLookupKey(tenantID, implID string) string {
	absPath := []string{implID}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func ccaTaLookupKey(tenantID, implID, instID string) string {
	absPath := []string{implID, instID}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func MustImplIDString(c psatoken.IClaims) string {
	v, err := c.GetImplID()
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(v)
}

func MustInstIDString(c psatoken.IClaims) string {
	v, err := c.GetInstID()
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(v)
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"scheme": &scheme.Plugin{
			Impl: &Scheme{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
