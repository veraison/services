// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

var testNonce = []byte{
	0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00,
	0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08,
	0x17, 0x16, 0x15, 0x14, 0x13, 0x12, 0x11, 0x10,
	0x1f, 0x1e, 0x1d, 0x1c, 0x1b, 0x1a, 0x19, 0x18,
}

func Test_ExtractClaims_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
		Nonce:    testNonce,
	}
	ta := string(taEndValBytes)
	claims, err := handler.ExtractClaims(&token, []string{ta})

	require.NoError(t, err)
	assert.Equal(t, "http://amd.com/2024/snp-corim-profile", claims["profile"].(string))
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
		Nonce:    testNonce,
	}

	ta := string(taEndValBytes)
	err = handler.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

	assert.NoError(t, err)
}

func Test_ValidateEvidenceIntegrity_BadTA(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-token.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement-bad.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
		Nonce:    testNonce,
	}

	ta := string(taEndValBytes)
	err = handler.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

	assert.EqualError(t, err, "ARK in evidence does not match provisioned ARK")
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	endorsementsBytes, err := os.ReadFile("test/refval-endorsement.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	result, err := handler.AppraiseEvidence(&ec, []string{string(endorsementsBytes)})
	require.NoError(t, err)

	attestation := result.Submods["SEVSNP"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
}
