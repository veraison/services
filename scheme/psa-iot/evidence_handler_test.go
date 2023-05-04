// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

func Test_GetTrustAnchorID_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/psa-token.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	expectedTaID := "PSA_IOT://1/BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg=/AQcGBQQDAgEADw4NDAsKCQgXFhUUExIREB8eHRwbGhkY"

	handler := &EvidenceHandler{}

	taID, err := handler.GetTrustAnchorID(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

func Test_ExtractVerifiedClaimsInteg_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/psaintegtoken.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-integ-endorsements.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
	}

	_, err = handler.ExtractClaims(&token, string(taEndValBytes))

	require.NoError(t, err)

}

func Test_ExtractVerifiedClaims_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/psa-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	extracted, err := handler.ExtractClaims(&token, string(taEndValBytes))

	require.NoError(t, err)
	assert.Equal(t, "PSA_IOT_PROFILE_1", extracted.ClaimsSet["psa-profile"].(string))

	swComponents := extracted.ClaimsSet["psa-software-components"].([]interface{})
	assert.Len(t, swComponents, 4)
	assert.Equal(t, "BL", swComponents[0].(map[string]interface{})["measurement-type"].(string))
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/psa-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	err = handler.ValidateEvidenceIntegrity(&token, string(taEndValBytes), nil)

	assert.NoError(t, err)
}

func Test_ValidateEvidenceIntegrity_BadKey(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "invalid public key",
			input:       "test/ta-bad-key.json",
			expectedErr: `could not get public key from trust anchor: unable to parse public key: asn1: structure error: tags don't match (16 vs {class:0 tag:2 length:1 isCompound:false}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} AlgorithmIdentifier @2`,
		},
		{
			desc:        "bad pem key header",
			input:       "test/ta-bad-key-header.json",
			expectedErr: `could not get public key from trust anchor: could not extract trust anchor PEM block`,
		},
		{
			desc:        "incorrect key type",
			input:       "test/ta-bad-key-private-key.json",
			expectedErr: "could not get public key from trust anchor: unsupported key type: \"PRIVATE KEY\"",
		},
	}

	for _, tv := range tvs {
		tokenBytes, err := os.ReadFile("test/psa-token.cbor")
		require.NoError(t, err)

		taEndValBytes, err := os.ReadFile(tv.input)
		require.NoError(t, err)
		h := &EvidenceHandler{}

		token := proto.AttestationToken{
			TenantId: "1",
			Data:     tokenBytes,
		}

		err = h.ValidateEvidenceIntegrity(&token, string(taEndValBytes), nil)
		assert.EqualError(t, err, tv.expectedErr)
	}
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	endorsementsBytes, err := os.ReadFile("test/endorsements.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	result, err := handler.AppraiseEvidence(&ec, []string{string(endorsementsBytes)})
	require.NoError(t, err)

	attestation := result.Submods["PSA_IOT"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
}

func Test_AppraiseEvidenceMultEndorsement_ok(t *testing.T) {
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	var endorsemementsArray []string
	endorsementsBytes, err := os.ReadFile("test/mult-endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	result, err := handler.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)

	attestation := result.Submods["PSA_IOT"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
}
