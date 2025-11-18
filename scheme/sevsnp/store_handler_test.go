// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	"encoding/json"
	"github.com/veraison/services/proto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/handler"
)

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	var e handler.Endorsement

	endorsementsBytes, err := os.ReadFile("test/refval-endorsement.json")
	require.NoError(t, err)

	err = json.Unmarshal(endorsementsBytes, &e)
	require.NoError(t, err)
	expectedKey := "SEVSNP://0/7699e6ac12ccdfd1dfac70e649ce1f046cb2afbb003438f4cdddfe2ccbe182fa5ffbe8dcdb930454324e10c52c788980"

	scheme := &StoreHandler{}
	keys, err := scheme.SynthKeysFromRefValue("0", &e)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, keys[0])

}

func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	var e handler.Endorsement

	endorsementsBytes, err := os.ReadFile("test/ta-endorsement.json")
	require.NoError(t, err)

	err = json.Unmarshal(endorsementsBytes, &e)
	require.NoError(t, err)

	expectedKey := "SEVSNP://ARK-Genoa"

	scheme := &StoreHandler{}
	keys, err := scheme.SynthKeysFromTrustAnchor("0", &e)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, keys[0])

}

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-ratsd-token")
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId:  "0",
		Data:      tokenBytes,
		MediaType: EvidenceMediaTypeRATSd,
		Nonce:     testNonce,
	}

	expectedTaID := "SEVSNP://ARK-Genoa"

	handler := &StoreHandler{}

	taIDs, err := handler.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, 1, len(taIDs))
	assert.Equal(t, expectedTaID, taIDs[0])
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	tokenBytes, err := os.ReadFile("test/sevsnp-ratsd-token")
	require.NoError(t, err)

	taEndValBytes, err := os.ReadFile("test/ta-endorsement.json")
	require.NoError(t, err)

	handler := &EvidenceHandler{}

	token := proto.AttestationToken{
		TenantId:  "0",
		Data:      tokenBytes,
		MediaType: EvidenceMediaTypeRATSd,
		Nonce:     testNonce,
	}
	ta := string(taEndValBytes)
	claims, err := handler.ExtractClaims(&token, []string{ta})
	require.NoError(t, err)

	expectedRefvalIDs := []string{"SEVSNP://0/7699e6ac12ccdfd1dfac70e649ce1f046cb2afbb003438f4cdddfe2ccbe182fa5ffbe8dcdb930454324e10c52c788980"}

	scheme := &StoreHandler{}
	refvalIDs, err := scheme.GetRefValueIDs("0", nil, claims)
	require.NoError(t, err)
	assert.Equal(t, expectedRefvalIDs, refvalIDs)
}
