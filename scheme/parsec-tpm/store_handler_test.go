// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

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

	expectedTaID := "PARSEC_TPM://1/AYiFVFnuuemzSkbrSMs58vaqadoEUioybRI9XFAfziEM"

	handler := &StoreHandler{}

	taIDs, err := handler.GetTrustAnchorIDs(&token)
	require.NoError(t, err)
	assert.Equal(t, []string{expectedTaID}, taIDs)
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	rawTA, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)

	trustAnchors := []string{string(rawTA)}

	rawToken, err := os.ReadFile("test/evidence/extracted.json")
	require.NoError(t, err)

	tokenJSON := make(map[string]interface{})
	err = json.Unmarshal(rawToken, &tokenJSON)
	require.NoError(t, err)

	claims := tokenJSON["evidence"].(map[string]interface{})

	expectedRefvalID := "PARSEC_TPM://1/cd1f0e55-26f9-460d-b9d8-f7fde171787c"

	handler := &StoreHandler{}

	refvalIDs, err := handler.GetRefValueIDs("1", trustAnchors, claims)
	require.NoError(t, err)
	assert.Equal(t, []string{expectedRefvalID}, refvalIDs)
}


func Test_SynthKeysFromTrustAnchor_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/evidence/ta_endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "PARSEC_TPM://1/AagIEsUMYDNxd1p5UuAACkxJGfJf9rcUZ/oyRFHDcAxn"

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromTrustAnchor("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])

}

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/evidence/refval-endorsements.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "PARSEC_TPM://1/cd1f0e55-26f9-460d-b9d8-f7fde171787c"

	scheme := &StoreHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}
