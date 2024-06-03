// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package psa_iot

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
	tokenBytes, err := os.ReadFile("test/psa-token.cbor")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Data:     tokenBytes,
		Nonce:    testNonce,
	}

	expectedTaID := "PSA_IOT://1/BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg=/AQcGBQQDAgEADw4NDAsKCQgXFhUUExIREB8eHRwbGhkY"

	handler := &StoreHandler{}

	taIDs, err := handler.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, 1, len(taIDs))
	assert.Equal(t, expectedTaID, taIDs[0])
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	rawToken, err := os.ReadFile("test/psa-token.json")
	require.NoError(t, err)

	claims := make(map[string]interface{})
	err = json.Unmarshal(rawToken, &claims)
	require.NoError(t, err)


	expectedRefvalIDs := []string{"PSA_IOT://1/BwYFBAMCAQAPDg0MCwoJCBcWFRQTEhEQHx4dHBsaGRg="}

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
	expectedKey := "PSA_IOT://1/76543210fedcba9817161514131211101f1e1d1c1b1a1918/Ac7rrnuJJ6MiflMDz14PH3s0u1Qq1yUKwD+83jbsLxUI"

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromTrustAnchor("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])

}

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "PSA_IOT://1/76543210fedcba9817161514131211101f1e1d1c1b1a1918"

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}
