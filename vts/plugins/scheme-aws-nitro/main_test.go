// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/veraison/services/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTime time.Time = time.Date(2022, 11, 9, 23, 0, 0, 0, time.UTC)

func Test_GetTrustAnchorID_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/aws_nitro_document.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_PSA_IOT,
		Data:     tokenBytes,
	}

	expectedTaID := "AWS_NITRO://1/"

	scheme := &Scheme{}

	taID, err := scheme.GetTrustAnchorID(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

// func Test_ExtractVerifiedClaimsInteg_ok(t *testing.T) {
// 	tokenBytes, err := os.ReadFile("test/psaintegtoken.cbor")
// 	require.NoError(t, err)

// 	taEndValBytes, err := os.ReadFile("test/ta-integ-endorsements.json")
// 	require.NoError(t, err)

// 	scheme := &Scheme{}

// 	token := proto.AttestationToken{
// 		TenantId: "0",
// 		Format:   proto.AttestationFormat_PSA_IOT,
// 		Data:     tokenBytes,
// 	}

// 	_, err = scheme.ExtractClaims(&token, string(taEndValBytes))

// 	require.NoError(t, err)

// }

func Test_ExtractVerifiedClaims_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/aws_nitro_document.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	extracted, err := scheme.ExtractClaimsTest(&token, string(taEndValBytes), testTime)

	require.NoError(t, err)
	expectedPcr0 := [48]byte{
		34, 249, 225, 201, 73, 32, 141, 165, 94, 176, 27, 155, 159, 200, 143, 135,
		69, 79, 119, 186, 19, 63, 13, 130, 50, 11, 80, 150, 33, 201, 36, 130,
		21, 42, 153, 208, 161, 35, 53, 185, 113, 120, 192, 45, 111, 151, 125, 1,
	}
	assert.Equal(t, expectedPcr0[:], extracted.ClaimsSet["PCR0"].([]byte))

	expectedNonce := [32]byte{
		198, 120, 200, 97, 53, 222, 83, 157, 24, 58, 207, 245, 136, 134, 217, 141,
		251, 152, 35, 4, 26, 249, 249, 52, 191, 144, 154, 192, 248, 217, 98, 69,
	}
	nonce := extracted.ClaimsSet["nonce"].([]byte)
	assert.Equal(t, expectedNonce[:], nonce)

	expectedUserData := [32]byte{
		124, 55, 16, 128, 121, 179, 232, 163, 109, 138, 121, 112, 222, 29, 109, 79,
		241, 70, 30, 14, 53, 217, 85, 124, 77, 120, 157, 245, 224, 87, 102, 32,
	}
	user_data := extracted.ClaimsSet["user_data"].([]byte)
	assert.Equal(t, expectedUserData[:], user_data)
}

func Test_ExtractVerifiedClaims_bad_signature(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/aws_nitro_document_bad_sig.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	_, err = scheme.ExtractClaimsTest(&token, string(taEndValBytes), testTime)

	assert.EqualError(t, err, `scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to AuthenticateDocument failed:AuthenticateDocument::Verify failed:verification error`)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/aws_nitro_document.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	err = scheme.ValidateEvidenceIntegrityTest(&token, string(taEndValBytes), nil, testTime)

	assert.NoError(t, err)
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	endorsementsBytes, err := os.ReadFile("test/endorsements.json")
	require.NoError(t, err)

	scheme := &Scheme{}

	attestation, err := scheme.AppraiseEvidence(&ec, []string{string(endorsementsBytes)})
	require.NoError(t, err)

	assert.Equal(t, proto.TrustTier_AFFIRMING, attestation.Result.Status)
}
