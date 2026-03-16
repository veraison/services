// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/vts/appraisal"
)

//go:embed test/evidence/sevsnp-ratsd-token
var sevsnpRatsdToken []byte

//go:embed test/evidence/sevsnp-ratsd-claims.json
var sevsnpRatsdClaimsMapJson []byte

//go:embed test/evidence/sevsnp-ratsd-env.json
var sevsnpRatsdEnv []byte

//go:embed test/corim/corim-sevsnp-valid.cbor
var sevsnpCorimValidEndorsements []byte

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	Impl := NewImplementation()
	evidence := appraisal.Evidence{Data: sevsnpRatsdToken, MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`}
	env, err := Impl.GetTrustAnchorIDs(&evidence)
	require.NoError(t, err)

	expectedVendor := "Advanced Micro Devices"
	expectedModel := "ARK-Genoa"
	expectedEnv := comid.Environment{Class: &comid.Class{Vendor: &expectedVendor, Model: &expectedModel}}

	require.Len(t, env, 1)
	assert.Equal(t, *env[0], expectedEnv)
}

func Test_GetReferenceValueIDs_ok(t *testing.T) {
	var (
		claims      map[string]any
		expectedEnv comid.Environment
	)

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &claims)
	require.NoError(t, err)

	Impl := NewImplementation()
	env, err := Impl.GetReferenceValueIDs(nil, claims)
	require.NoError(t, err)
	require.Len(t, env, 1)

	err = json.Unmarshal(sevsnpRatsdEnv, &expectedEnv)
	require.NoError(t, err)

	assert.Equal(t, *env[0], expectedEnv)
}

func Test_ExtractClaims_ok(t *testing.T) {
	var expectedClaims map[string]any

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &expectedClaims)
	require.NoError(t, err)

	expectedMeasurementMap, err := transformClaimsToMeasurementsMap(expectedClaims)
	require.NoError(t, err)

	Impl := NewImplementation()
	evidence := appraisal.Evidence{Data: sevsnpRatsdToken, MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`}
	claims, err := Impl.ExtractClaims(&evidence, nil)
	require.NoError(t, err)
	measurementMap, err := transformClaimsToMeasurementsMap(claims)
	require.NoError(t, err)

	assert.Equal(t, expectedMeasurementMap, measurementMap)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	var (
		endCorim     corim.UnsignedCorim
		endComid     comid.Comid
		trustAnchors []*comid.KeyTriple
	)

	err := endCorim.FromCBOR(sevsnpCorimValidEndorsements)
	require.NoError(t, err)
	require.Len(t, endCorim.Tags, 2)
	err = endComid.FromCBOR(endCorim.Tags[1].Content)
	require.NoError(t, err)
	keyTriples := *endComid.Triples.AttestVerifKeys
	trustAnchors = append(trustAnchors, &keyTriples[0])

	nonce, err := hex.DecodeString("4d4944424e48323869696f69736a5079787878787878787878787878787878784d4944424e48323869696f69736a507978787878787878787878787878787878")
	require.NoError(t, err)

	incorrectNonce := make([]byte, len(nonce))
	copy(incorrectNonce, nonce)
	incorrectNonce[0] ^= 0xff

	Impl := NewImplementation()

	testCases := []struct {
		name    string
		nonce   []byte
		wantErr bool
	}{
		{name: "valid nonce", nonce: nonce},
		{name: "invalid nonce", nonce: incorrectNonce, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evidence := appraisal.Evidence{
				Data:      sevsnpRatsdToken,
				MediaType: `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`,
				Nonce:     tc.nonce,
			}

			err := Impl.ValidateEvidenceIntegrity(&evidence, trustAnchors, nil)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_AppraiseClaims_ok(t *testing.T) {
	var (
		claims       map[string]any
		endCorim     corim.UnsignedCorim
		endComid     comid.Comid
		endorsements []*comid.ValueTriple
	)

	err := json.Unmarshal(sevsnpRatsdClaimsMapJson, &claims)
	require.NoError(t, err)

	err = endCorim.FromCBOR(sevsnpCorimValidEndorsements)
	require.NoError(t, err)
	require.Len(t, endCorim.Tags, 2)
	err = endComid.FromCBOR(endCorim.Tags[0].Content)
	require.NoError(t, err)

	Impl := NewImplementation()
	endorsements = append(endorsements, &endComid.Triples.ReferenceValues.Values[0])
	attestationResult, err := Impl.AppraiseClaims(claims, endorsements)
	require.NoError(t, err)

	sevsnpSubmod := attestationResult.Submods["SEVSNP"]

	assert.Equal(t, ear.TrustTierAffirming, *sevsnpSubmod.Status)
	assert.Equal(t, ear.GenuineHardwareClaim, sevsnpSubmod.TrustVector.Hardware)
	assert.Equal(t, ear.EncryptedMemoryRuntimeClaim, sevsnpSubmod.TrustVector.RuntimeOpaque)
}
