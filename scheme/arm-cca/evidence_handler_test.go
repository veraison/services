// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

func Test_AppraiseEvidence_Platform_ok(t *testing.T) { // nolint: dupl
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

func Test_AppraiseEvidence_Realm(t *testing.T) { // nolint: dupl
	tvs := []struct {
		desc           string
		input          string
		expectedStatus ear.TrustTier
		expectedExec   ear.TrustClaim
	}{

		{
			desc:           "No realm endorsements",
			input:          "test/no-realm-endorsements.json",
			expectedStatus: ear.TrustTierWarning,
			expectedExec:   ear.UnrecognizedRuntimeClaim,
		},
		{
			desc:           "No matching rim measurements",
			input:          "test/rim-mismatch-endorsements.json",
			expectedStatus: ear.TrustTierContraindicated,
			expectedExec:   ear.ContraindicatedRuntimeClaim,
		},
		{
			desc:           "matching rim & rpv, no rem",
			input:          "test/no-rem-endorsements.json",
			expectedStatus: ear.TrustTierAffirming,
			expectedExec:   ear.ApprovedBootClaim,
		},
		{
			desc:           "matching rim & rem, no rpv",
			input:          "test/no-rpv-endorsements.json",
			expectedStatus: ear.TrustTierAffirming,
			expectedExec:   ear.ApprovedRuntimeClaim,
		},
		{
			desc:           "matching rim, rpv and rem measurements",
			input:          "test/match-endorsements.json",
			expectedStatus: ear.TrustTierAffirming,
			expectedExec:   ear.ApprovedRuntimeClaim,
		},
	}
	for _, tv := range tvs {
		extractedBytes, err := os.ReadFile("test/extracted.json")
		require.NoError(t, err)

		var ec proto.EvidenceContext
		err = json.Unmarshal(extractedBytes, &ec)
		require.NoError(t, err)
		var endorsemementsArray []string
		endorsementsBytes, err := os.ReadFile(tv.input)
		require.NoError(t, err)
		err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
		require.NoError(t, err)
		scheme := &EvidenceHandler{}
		result, err := scheme.AppraiseEvidence(&ec, endorsemementsArray)
		require.NoError(t, err)

		attestation := result.Submods["CCA_REALM"]
		assert.Equal(t, tv.expectedStatus, *attestation.Status)
		assert.Equal(t, tv.expectedExec, attestation.TrustVector.Executables)
	}
}

func Test_AppraiseEvidence_Platform_mismatch_refval_meas(t *testing.T) { // nolint: dupl
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

func Test_AppraiseEvidence_Platform_mismatch_refval_cfg(t *testing.T) { // nolint: dupl
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
