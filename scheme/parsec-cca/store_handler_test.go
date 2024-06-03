// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/evidence/evidence.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
	}

	expectedTaID := "PARSEC_CCA://1/f0VMRgIBAQAAAAAAAAAAAAMAPgABAAAAUFgAAAAAAAA=/AQcGBQQDAgEADw4NDAsKCQgXFhUUExIREB8eHRwbGhkY"

	handler := &StoreHandler{}

	taIDs, err := handler.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taIDs[0])
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	rawToken, err := os.ReadFile("test/evidence/extracted.json")
	require.NoError(t, err)

	tokenJSON := make(map[string]interface{})
	err = json.Unmarshal(rawToken, &tokenJSON)
	require.NoError(t, err)

	claims := tokenJSON["evidence"].(map[string]interface{})

	expectedRefvalIDs := []string{"PARSEC_CCA://1/f0VMRgIBAQAAAAAAAAAAAAMAPgABAAAAUFgAAAAAAAA="}

	scheme := &StoreHandler{}
	refvalIDs, err := scheme.GetRefValueIDs("1", nil, claims)
	require.NoError(t, err)
	assert.Equal(t, expectedRefvalIDs, refvalIDs)
}

func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "PARSEC_CCA://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=/Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromTrustAnchor("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])

}

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/evidence/refval_endorsement.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "PARSEC_CCA://1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}
