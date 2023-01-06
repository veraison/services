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
	"github.com/veraison/ear"
	"github.com/veraison/psatoken"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"github.com/veraison/services/vts/plugins/common"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type SwAttr struct {
	ImplID    *[]byte `json:"psa.impl-id"`
	Model     *string `json:"psa.hw-model"`
	Vendor    *string `json:"psa.hw-vendor"`
	MeasDesc  *uint64 `json:"psa.measurement-desc"`
	MeasType  *string `json:"psa.measurement-type"`
	MeasValue *[]byte `json:"psa.measurement-value"`
	SignerID  *[]byte `json:"psa.signer-id"`
	Version   *string `json:"psa.version"`
}

type Endorsements struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Attr   SwAttr `json:"attributes"`
}

type TaAttr struct {
	Model    *string `json:"psa.hw-model"`
	Vendor   *string `json:"psa.hw-vendor"`
	VerifKey *string `json:"psa.iak-pub"`
	ImplID   *[]byte `json:"psa.impl-id"`
	InstID   *string `json:"psa.inst-id"`
}

type TaEndorsements struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`
	Attr   TaAttr `json:"attributes"`
}

const SchemeName = "PSA_IOT"

type Scheme struct{}

func (s Scheme) GetName() string {
	return SchemeName
}

func (s Scheme) SynthKeysFromSwComponent(
	tenantID string,
	swComp *proto.Endorsement,
) ([]string, error) {
	var (
		implID string
		fields map[string]*structpb.Value
		err    error
	)
	log.Printf("SynthKeysFromSwComponent called\n")
	fields, err = common.GetFieldsFromParts(swComp.GetAttributes())
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize software component abs-path: %w", err)
	}

	implID, err = common.GetMandatoryPathSegment("psa.impl-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize software component abs-path: %w", err)
	}

	finalstr := psaSoftwareLookupKey(tenantID, implID)
	log.Printf("PSA Plugin PSA Look Up Key= %s\n", finalstr)
	return []string{psaSoftwareLookupKey(tenantID, implID)}, nil
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

	implID, err = common.GetMandatoryPathSegment("psa.impl-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	instID, err = common.GetMandatoryPathSegment("psa.inst-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize trust anchor abs-path: %w", err)
	}

	finalstr := psaTaLookupKey(tenantID, implID, instID)
	log.Printf("PSA Plugin TA PSA Look Up Key= %s\n", finalstr)
	return []string{psaTaLookupKey(tenantID, implID, instID)}, nil
}

func (s Scheme) GetSupportedMediaTypes() []string {
	return []string{
		"application/psa-attestation-token",
		"application/eat-cwt; profile=http://arm.com/psa/2.0.0",
	}
}

func (s Scheme) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	var psaToken psatoken.Evidence

	err := psaToken.FromCOSE(token.Data)
	if err != nil {
		return "", err
	}

	return psaTaLookupKey(
		token.TenantId,
		MustImplIDString(psaToken.Claims),
		MustInstIDString(psaToken.Claims),
	), nil
}

func (s Scheme) ExtractClaims(
	token *proto.AttestationToken,
	trustAnchor string,
) (*scheme.ExtractedClaims, error) {
	var psaToken psatoken.Evidence

	if err := psaToken.FromCOSE(token.Data); err != nil {
		return nil, err
	}

	var extracted scheme.ExtractedClaims

	claimsSet, err := claimsToMap(psaToken.Claims)
	if err != nil {
		return nil, err
	}
	extracted.ClaimsSet = claimsSet

	extracted.SoftwareID = psaSoftwareLookupKey(
		token.TenantId,
		MustImplIDString(psaToken.Claims),
	)
	log.Printf("\n Extracted SW ID Key = %s", extracted.SoftwareID)
	return &extracted, nil
}

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
	ta := *endorsement.Attr.VerifKey
	block, rest := pem.Decode([]byte(ta))

	if block == nil {
		log.Println("Could not get TA PEM Block ExtractVerifiedClaims")
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

	var psaToken psatoken.Evidence

	if err = psaToken.FromCOSE(token.Data); err != nil {
		return err
	}

	if err = psaToken.Verify(pk); err != nil {
		return err
	}
	log.Println("\n Token Signature Verified")
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

	return result, err
}

func claimsToMap(claims psatoken.IClaims) (map[string]interface{}, error) {
	data, err := claims.ToJSON()
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

	match := matchSoftware(claims, endorsements)
	if match {
		result.TrustVector.Executables = ear.ApprovedRuntimeClaim
		log.Println("\n matchSoftware Success")

	} else {
		result.TrustVector.Executables = ear.UnrecognizedRuntimeClaim
		log.Println("\n matchSoftware Failed")
	}

	result.UpdateStatusFromTrustVector()

	result.VeraisonProcessedEvidence = &evidence

	return nil
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
		key := base64.StdEncoding.EncodeToString(*endorsement.Attr.MeasValue)
		evComp, ok := evidenceComponents[key]
		if !ok {
			matched = false
			break
		}

		log.Printf("MeasType Evidence: %s, Endorsement: %s", *evComp.MeasurementType, *endorsement.Attr.MeasType)
		typeMatched := *endorsement.Attr.MeasType == "" || *endorsement.Attr.MeasType == *evComp.MeasurementType
		sigMatched := *endorsement.Attr.SignerID == nil || bytes.Equal(*endorsement.Attr.SignerID, *evComp.SignerID)
		versionMatched := *endorsement.Attr.Version == "" || *endorsement.Attr.Version == *evComp.Version

		if !(typeMatched && sigMatched && versionMatched) {
			matched = false
			break
		}
	}
	return matched
}

func psaSoftwareLookupKey(tenantID, implID string) string {
	absPath := []string{implID}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func psaTaLookupKey(tenantID, implID, instID string) string {
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
