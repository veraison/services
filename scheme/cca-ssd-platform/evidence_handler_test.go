// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_ssd_platform

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

func Test_GetTrustAnchorID_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/cca-token.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	expectedTaID := "CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/AQICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIC"

	scheme := &EvidenceHandler{}

	taID, err := scheme.GetTrustAnchorID(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"

	scheme := &EvidenceHandler{}
	key_list, err := scheme.SynthKeysFromTrustAnchor("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])

}

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/refval-endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	scheme := &EvidenceHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}

func Test_AppraiseEvidence_ok(t *testing.T) { // nolint: dupl
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	var endorsemementsArray []string
	endorsementsBytes, err := os.ReadFile("test/endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	result, err := scheme.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)

	attestation := result.Submods["CCA_SSD_PLATFORM"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
	assert.Equal(t, attestation.TrustVector.Executables, ear.ApprovedRuntimeClaim)
	assert.Equal(t, attestation.TrustVector.Configuration, ear.ApprovedConfigClaim)
}

func Test_AppraiseEvidence_mismatch_refval_meas(t *testing.T) { // nolint: dupl
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	var endorsemementsArray []string
	endorsementsBytes, err := os.ReadFile("test/mismatch-refval-endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	result, err := scheme.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)

	attestation := result.Submods["CCA_SSD_PLATFORM"]

	assert.Equal(t, ear.TrustTierWarning, *attestation.Status)
	assert.Equal(t, attestation.TrustVector.Executables, ear.UnrecognizedRuntimeClaim)
	assert.Equal(t, attestation.TrustVector.Configuration, ear.ApprovedConfigClaim)
}

func Test_AppraiseEvidence_mismatch_refval_cfg(t *testing.T) { // nolint: dupl
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	var endorsemementsArray []string
	endorsementsBytes, err := os.ReadFile("test/mismatch-cfg-endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	result, err := scheme.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)

	attestation := result.Submods["CCA_SSD_PLATFORM"]

	assert.Equal(t, ear.TrustTierWarning, *attestation.Status)
	assert.Equal(t, attestation.TrustVector.Executables, ear.ApprovedRuntimeClaim)
	assert.Equal(t, attestation.TrustVector.Configuration, ear.UnsafeConfigClaim)
}

func Test_ExtractVerifiedClaims_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/cca-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	extracted, err := scheme.ExtractClaims(&token, string(taEndValBytes))

	require.NoError(t, err)
	assert.Equal(t, "http://arm.com/CCA-SSD/1.0.0", extracted.ClaimsSet["cca-platform-profile"].(string))

	swComponents := extracted.ClaimsSet["cca-platform-sw-components"].([]interface{})
	assert.Len(t, swComponents, 4)
	assert.Equal(t, "BL", swComponents[0].(map[string]interface{})["measurement-type"].(string))
	ccaPlatformCfg := extracted.ClaimsSet["cca-platform-config"]
	assert.Equal(t, "AQID", ccaPlatformCfg)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/cca-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	err = scheme.ValidateEvidenceIntegrity(&token, string(taEndValBytes), nil)

	assert.NoError(t, err)
}

func Test_ValidateEvidenceIntegrity_invalid_key(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/cca-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/invalid-key-ta-endorsements.json")
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}
	expectedErr := `could not get public key from trust anchor: unsupported key type: "PRIVATE KEY"`

	err = scheme.ValidateEvidenceIntegrity(&token, string(taEndValBytes), nil)
	assert.EqualError(t, err, expectedErr)
}

func Test_GetSupportedMediaType_ok(t *testing.T) {
	expectedMt := "application/eat-collection; profile=http://arm.com/CCA-SSD/1.0.0"
	scheme := &EvidenceHandler{}
	mtList := scheme.GetSupportedMediaTypes()
	assert.Len(t, mtList, 1)
	assert.Equal(t, mtList[0], expectedMt)
}
