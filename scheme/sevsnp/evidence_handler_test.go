// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/ear"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/veraison/services/proto"
)

var testNonce = []byte{
	77, 73, 68, 66, 78, 72, 50, 56,
	105, 105, 111, 105, 115, 106, 80, 121,
	120, 120, 120, 120, 120, 120, 120, 120,
	120, 120, 120, 120, 120, 120, 120, 120,
	77, 73, 68, 66, 78, 72, 50, 56,
	105, 105, 111, 105, 115, 106, 80, 121,
	120, 120, 120, 120, 120, 120, 120, 120,
	120, 120, 120, 120, 120, 120, 120, 120,
}

var testBadNonce = []byte{
	0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00,
	0x0f, 0x0e, 0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08,
	0x17, 0x16, 0x15, 0x14, 0x13, 0x12, 0x11, 0x10,
	0x1f, 0x1e, 0x1d, 0x1c, 0x1b, 0x1a, 0x19, 0x18,
}

type sevSnpEvidence struct {
	FileName  string
	MediaType string
}

var testEvidenceList = []sevSnpEvidence{
	{"test/sevsnp-ratsd-token", EvidenceMediaTypeRATSd},
	{"test/sevsnp-tsm-report.json", EvidenceMediaTypeTSMJson},
	{"test/sevsnp-tsm-report.cbor", EvidenceMediaTypeTSMCbor},
}

func Test_ExtractClaims_ok(t *testing.T) {
	for _, evidence := range testEvidenceList {
		tokenBytes, err := os.ReadFile(evidence.FileName)
		require.NoError(t, err)

		taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
		require.NoError(t, err)

		handler := &EvidenceHandler{}

		token := proto.AttestationToken{
			TenantId:  "0",
			Data:      tokenBytes,
			MediaType: evidence.MediaType,
			Nonce:     testNonce,
		}
		ta := string(taEndValBytes)
		_, err = handler.ExtractClaims(&token, []string{ta})

		require.NoError(t, err)
	}
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	for _, evidence := range testEvidenceList {
		tokenBytes, err := os.ReadFile(evidence.FileName)
		require.NoError(t, err)

		taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
		require.NoError(t, err)

		handler := &EvidenceHandler{}

		token := proto.AttestationToken{
			TenantId:  "0",
			Data:      tokenBytes,
			MediaType: evidence.MediaType,
			Nonce:     testNonce,
		}

		ta := string(taEndValBytes)
		err = handler.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

		assert.NoError(t, err)
	}
}

func Test_ValidateEvidenceIntegrity_BadTA(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-ratsd-token")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement-bad.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId:  "0",
		Data:      tokenBytes,
		MediaType: EvidenceMediaTypeRATSd,
		Nonce:     testNonce,
	}

	ta := string(taEndValBytes)
	err = handler.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

	assert.EqualError(t, err, "{\"detail\":[\"evidence Trust Anchor (ARK) doesn't match the provisioned one\"],\"detail-type\":\"error\",\"error\":\"bad evidence\"}")
}

func Test_ValidateEvidenceIntegrity_BadNonce(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-ratsd-token")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId:  "0",
		Data:      tokenBytes,
		MediaType: EvidenceMediaTypeRATSd,
		Nonce:     testBadNonce,
	}

	ta := string(taEndValBytes)
	err = handler.ValidateEvidenceIntegrity(&token, []string{ta}, nil)

	assert.EqualError(t, err, "{\"detail\":[\"nonce in the evidence doesn't match the session nonce. evidence: 0x4d4944424e48323869696f69736a5079787878787878787878787878787878784d4944424e48323869696f69736a507978787878787878787878787878787878 vs session: 0x07060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918\"],\"detail-type\":\"error\",\"error\":\"bad evidence\"}")
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	for _, evidence := range testEvidenceList {
		tokenBytes, err := os.ReadFile(evidence.FileName)
		require.NoError(t, err)

		taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
		require.NoError(t, err)

		handler := &EvidenceHandler{}

		token := proto.AttestationToken{
			TenantId:  "0",
			Data:      tokenBytes,
			MediaType: evidence.MediaType,
			Nonce:     testNonce,
		}
		ta := string(taEndValBytes)
		claims, err := handler.ExtractClaims(&token, []string{ta})
		require.NoError(t, err)

		claimsJson, err := json.Marshal(claims)
		require.NoError(t, err)

		var ec proto.EvidenceContext
		ec.TenantId = "0"
		ec.TrustAnchorIds = []string{"SEVSNP://ARK-Genoa"}
		ec.ReferenceIds = []string{"SEVSNP://0/7699e6ac12ccdfd1dfac70e649ce1f046cb2afbb003438f4cdddfe2ccbe182fa5ffbe8dcdb930454324e10c52c788980"}
		err = json.Unmarshal(claimsJson, &ec.Evidence)
		require.NoError(t, err)

		endorsementsBytes, err := os.ReadFile("test/refval-endorsement.json")
		require.NoError(t, err)

		result, err := handler.AppraiseEvidence(&ec, []string{string(endorsementsBytes)})
		require.NoError(t, err)

		attestation := result.Submods["SEVSNP"]

		assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
	}
}
