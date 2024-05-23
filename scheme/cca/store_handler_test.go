// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

var testNonce = []byte{
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
	0x41, 0x42, 0x41, 0x42, 0x41, 0x42, 0x41, 0x42,
}

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/cca-token.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
		Nonce:    testNonce,
	}

	expectedTaID := []string{"CCA_SSD://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/AQICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIC"}

	scheme := &StoreHandler{}

	taID, err := scheme.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "CCA_SSD://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"

	scheme := &StoreHandler{}
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
	expectedKey := "CCA_SSD://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}

func Test_GetReferenceIDs_ok(t *testing.T) {
	var ta []string
	var claims map[string]interface{}
	expectedRefValID := []string{
		"CCA_SSD://1/AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=",
		"CCA_SSD://1/Q0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQw==",
	}
	evidenceBytes, err := os.ReadFile("test/extracted-claims.json")
	require.NoError(t, err)
	err = json.Unmarshal(evidenceBytes, &claims)
	require.NoError(t, err)
	scheme := &StoreHandler{}
	refValID, err := scheme.GetRefValueIDs("1", ta, claims)
	require.NoError(t, err)
	assert.Equal(t, expectedRefValID, refValID)
}
