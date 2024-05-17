// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_realm_provisioning

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/handler"
)

func Test_GetName_ok(t *testing.T) {
	expectedName := "cca-realm-store-handler"
	handler := &StoreHandler{}
	name := handler.GetName()
	assert.Equal(t, expectedName, name)
}

func Test_GetAttestationScheme(t *testing.T) {
	expectedScheme := "CCA_REALM"
	handler := &StoreHandler{}
	name := handler.GetAttestationScheme()
	assert.Equal(t, expectedScheme, name)
}

func Test_GetSupportedMediaTypes(t *testing.T) {
	handler := &StoreHandler{}
	mt := handler.GetSupportedMediaTypes()
	assert.Nil(t, mt)
}

func Test_SynthKeysFromTrustAnchor_nok(t *testing.T) {
	scheme := &StoreHandler{}
	expectedErr := "unexpected SynthKeysFromTrustAnchor() invocation for scheme: CCA_REALM"
	_, err := scheme.SynthKeysFromTrustAnchor("1", nil)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErr)
}

func Test_GetTrustAnchorID_nok(t *testing.T) {
	scheme := &StoreHandler{}
	expectedErr := "unexpected GetTrustAnchorIDs() invocation for scheme: CCA_REALM"
	_, err := scheme.GetTrustAnchorIDs(nil)
	assert.NotNil(t, err)
	assert.EqualError(t, err, expectedErr)
}

func Test_SynthKeysFromRefValue1_ok(t *testing.T) {
	tvs := []struct {
		desc        string
		input       string
		expectedKey string
	}{
		{
			desc:        "no realm personalization in reference value",
			input:       "test/store/refvalEndorsementsNoRpv.json",
			expectedKey: "CCA_REALM://1/QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1",
		},
		{
			desc:        "complete reference value with rim and personalization value",
			input:       "test/store/refvalEndorsements.json",
			expectedKey: "CCA_REALM://1/QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1/5Fty9cDAtXLbTY06t+l/No/3TmI0eoJN7LZ6hOUiTXXkW3L1wMC1cttNjTq36X82j/dOYjR6gk3stnqE5SJNdQ==",
		},
	}

	for _, tv := range tvs {
		endorsementsBytes, err := os.ReadFile(tv.input)
		require.NoError(t, err)
		var endors handler.Endorsement
		err = json.Unmarshal(endorsementsBytes, &endors)
		require.NoError(t, err)
		scheme := &StoreHandler{}
		key, err := scheme.SynthKeysFromRefValue("1", &endors)
		require.NoError(t, err)
		assert.Equal(t, tv.expectedKey, key[0])
	}
}
