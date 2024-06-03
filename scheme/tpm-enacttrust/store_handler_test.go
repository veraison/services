// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/proto"
)

func Test_GetTrustAnchorIDs_ok(t *testing.T) {
	data, err := os.ReadFile("test/tokens/basic.token")
	require.NoError(t, err)

	ta := proto.AttestationToken{
		TenantId:  "0",
		MediaType: "application/vnd.enacttrust.tpm-evidence",
		Data:      data,
	}

	var s StoreHandler

	taIDs, err := s.GetTrustAnchorIDs(&ta)
	require.NoError(t, err)
	assert.Equal(t, "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1", taIDs[0])
}

func Test_GetRefValueIDs_ok(t *testing.T) {
	raw, err := os.ReadFile("test/tokens/basic.json")
	require.NoError(t, err)

	claims := make(map[string]interface{})
	err = json.Unmarshal(raw, &claims)
	require.NoError(t, err)

	expectedRefvalIDs := []string{"TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1"}

	s := StoreHandler{}

	refvalIDs, err := s.GetRefValueIDs("0", nil, claims)
	require.NoError(t, err)
	assert.Equal(t, expectedRefvalIDs, refvalIDs)
}
