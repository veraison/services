// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

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

	attestation := result.Submods["CCA"]

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

	attestation := result.Submods["CCA"]

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

	attestation := result.Submods["CCA"]

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
		Nonce:    testNonce,
	}
	ta := string(taEndValBytes)

	extracted, err := scheme.ExtractClaims(&token, []string{ta})
	platformClaims := extracted["platform"].(map[string]interface{})

	require.NoError(t, err)
	assert.Equal(t, "http://arm.com/CCA-SSD/1.0.0",
		platformClaims["cca-platform-profile"].(string))

	swComponents := platformClaims["cca-platform-sw-components"].([]interface{})
	assert.Len(t, swComponents, 4)
	assert.Equal(t, "BL", swComponents[0].(map[string]interface{})["measurement-type"].(string))
	ccaPlatformCfg := platformClaims["cca-platform-config"]
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
		Nonce:    testNonce,
	}
	ta := string(taEndValBytes)

	err = scheme.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

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
		Nonce:    testNonce,
	}
	expectedErr := `could not get public key from trust anchor: could not decode subject public key info: unsupported key type: "PRIVATE KEY"`

	ta := string(taEndValBytes)
	err = scheme.ValidateEvidenceIntegrity(&token, []string{ta}, nil)
	assert.EqualError(t, err, expectedErr)
}

func Test_GetSupportedMediaType_ok(t *testing.T) {
	expectedMt := "application/eat-collection; profile=http://arm.com/CCA-SSD/1.0.0"
	scheme := &EvidenceHandler{}
	mtList := scheme.GetSupportedMediaTypes()
	assert.Len(t, mtList, 1)
	assert.Equal(t, mtList[0], expectedMt)
}
