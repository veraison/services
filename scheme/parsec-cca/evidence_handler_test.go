// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

func Test_ExtractClaims_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/evidence/evidence.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
	}

	ta := string(taEndValBytes)
	_, err = handler.ExtractClaims(&token, []string{ta})
	require.NoError(t, err)
}

func Test_ExtractClaims_nok_bad_evidence(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/evidence/bad_evidence.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)
	expectedErr := "CBOR decoding: cbor: invalid additional information 28 for type byte string"
	h := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId: "0",
		Data:     tokenBytes,
	}
	ta := string(taEndValBytes)
	_, err = h.ExtractClaims(&token, []string{ta})
	err1 := errors.Unwrap(err)
	require.NotNil(t, err1)
	assert.EqualError(t, err1, expectedErr)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/evidence/evidence.cbor")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)
	h := &EvidenceHandler{}
	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}
	ta := string(taEndValBytes)
	err = h.ValidateEvidenceIntegrity(&token, []string{ta}, nil)
	require.NoError(t, err)
}

func Test_ValidateEvidenceIntegrity_nok(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "incorrect public key",
			input:       "test/evidence/unmatched_endorsements.json",
			expectedErr: `failed to verify signature: PAT validation failed: unable to verify platform token: verification error`,
		},
		{
			desc:        "invalid public key",
			input:       "test/evidence/bad_key_endorsements.json",
			expectedErr: `could not get public key from trust anchor: could not decode subject public key info: unable to parse public key: asn1: structure error: tags don't match (16 vs {class:0 tag:2 length:1 isCompound:false}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} AlgorithmIdentifier @2`,
		},
		{
			desc:        "bad pem key header",
			input:       "test/evidence/bad_key_header_endorsements.json",
			expectedErr: `could not get public key from trust anchor: could not decode subject public key info: could not extract trust anchor PEM block`,
		},
		{
			desc:        "incorrect key type",
			input:       "test/evidence/bad_key_private_key.json",
			expectedErr: "could not get public key from trust anchor: could not decode subject public key info: unsupported key type: \"PRIVATE KEY\"",
		},
	}

	for _, tv := range tvs {
		tokenBytes, err := os.ReadFile("test/evidence/evidence.cbor")
		require.NoError(t, err)

		taEndValBytes, err := os.ReadFile(tv.input)
		ta := string(taEndValBytes)
		require.NoError(t, err)
		h := &EvidenceHandler{}

		token := proto.AttestationToken{
			TenantId: "1",
			Data:     tokenBytes,
		}

		err = h.ValidateEvidenceIntegrity(&token, []string{ta}, nil)
		assert.EqualError(t, err, tv.expectedErr)
	}
}

func Test_AppraiseEvidence_mismatch(t *testing.T) {
	tvs := []struct {
		desc  string
		input string
	}{
		{
			desc:  "mismatch platform config",
			input: "test/evidence/mismatch_cfg_refval_endorsements.json",
		},
		{
			desc:  "mismatch SW Components",
			input: "test/evidence/mismatch_swcomp_refval_endorsements.json",
		},
	}

	for index, tv := range tvs {
		var endorsemementsArray []string
		extractedBytes, err := os.ReadFile("test/evidence/extracted.json")
		require.NoError(t, err)

		var ec proto.EvidenceContext
		err = json.Unmarshal(extractedBytes, &ec)
		require.NoError(t, err)
		endorsementsBytes, err := os.ReadFile(tv.input)
		require.NoError(t, err)
		err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
		require.NoError(t, err)

		handler := &EvidenceHandler{}

		result, err := handler.AppraiseEvidence(&ec, endorsemementsArray)
		require.NoError(t, err)
		attestation := result.Submods["PARSEC_CCA"]

		assert.Equal(t, ear.TrustTierWarning, *attestation.Status)
		if index == 0 {
			assert.Equal(t, attestation.TrustVector.Executables, ear.ApprovedRuntimeClaim)
			assert.Equal(t, attestation.TrustVector.Configuration, ear.UnsafeConfigClaim)
		} else {
			assert.Equal(t, attestation.TrustVector.Executables, ear.UnrecognizedRuntimeClaim)
			assert.Equal(t, attestation.TrustVector.Configuration, ear.ApprovedConfigClaim)
		}

	}
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	var endorsemementsArray []string
	extractedBytes, err := os.ReadFile("test/evidence/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)
	endorsementsBytes, err := os.ReadFile("test/evidence/refval_endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	result, err := handler.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)
	attestation := result.Submods["PARSEC_CCA"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
	assert.Equal(t, attestation.TrustVector.Executables, ear.ApprovedRuntimeClaim)
	assert.Equal(t, attestation.TrustVector.Configuration, ear.ApprovedConfigClaim)
}

func Test_GetName_ok(t *testing.T) {
	scheme := &EvidenceHandler{}
	expectedName := EvidenceHandlerName
	name := scheme.GetName()
	assert.Equal(t, name, expectedName)
}

func Test_GetAttestationScheme_ok(t *testing.T) {
	scheme := &EvidenceHandler{}
	expectedScheme := "PARSEC_CCA"
	name := scheme.GetAttestationScheme()
	assert.Equal(t, name, expectedScheme)
}

func Test_GetSupportedMediaTypes_ok(t *testing.T) {
	expectedMt := "application/vnd.parallaxsecond.key-attestation.cca"
	scheme := &EvidenceHandler{}
	mtList := scheme.GetSupportedMediaTypes()
	assert.Len(t, mtList, 1)
	assert.Equal(t, mtList[0], expectedMt)
}
