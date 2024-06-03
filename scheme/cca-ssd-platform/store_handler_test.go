// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_ssd_platform

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

	expectedTaID := []string{"CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/AQICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIC"}

	scheme := &StoreHandler{}

	taID, err := scheme.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	rawToken, err := os.ReadFile("test/cca-token.json")
	require.NoError(t, err)

	tokenJSON := make(map[string]map[string]interface{})
	err = json.Unmarshal(rawToken, &tokenJSON)
	require.NoError(t, err)

	platformClaims := tokenJSON["cca-platform-token"]
	platformClaims["cca-platform-challenge"] = testNonce

	claims := map[string]interface{}{"platform": platformClaims}

	expectedRefvalIDs := []string{"CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}

	scheme := &StoreHandler{}
	refvalIDs, err := scheme.GetRefValueIDs("1", nil, claims)
	require.NoError(t, err)
	assert.Equal(t, expectedRefvalIDs, refvalIDs)
}

func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/ta-endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"

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
	expectedKey := "CCA_SSD_PLATFORM://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}
