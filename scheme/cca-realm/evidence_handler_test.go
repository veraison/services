// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_realm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

func Test_SynthKeysFromRefValue_ok(t *testing.T) {
	endorsementsBytes, err := os.ReadFile("test/ref-val-endorsement.json")
	require.NoError(t, err)

	var endors handler.Endorsement
	err = json.Unmarshal(endorsementsBytes, &endors)
	require.NoError(t, err)
	expectedKey := "CCA_REALM://1/QoS1aUymwNLPR4mguVrIAlyBjeUjBDZL580pgbLS7caFsyInfsJYGZYkE9jJssH1"

	scheme := &EvidenceHandler{}
	key_list, err := scheme.SynthKeysFromRefValue("1", &endors)
	require.NoError(t, err)
	assert.Equal(t, expectedKey, key_list[0])
}

func Test_AppraiseEvidence_ok(t *testing.T) { // nolint: dupl
	extractedBytes, err := os.ReadFile("test/cca-claims.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	var endorsemementsArray []string
	endorsementsBytes, err := os.ReadFile("test/ref-val-endorsements.json")
	require.NoError(t, err)
	err = json.Unmarshal(endorsementsBytes, &endorsemementsArray)
	require.NoError(t, err)

	scheme := &EvidenceHandler{}

	result, err := scheme.AppraiseEvidence(&ec, endorsemementsArray)
	require.NoError(t, err)

	attestation := result.Submods["CCA_REALM"]

	assert.Equal(t, ear.TrustTierAffirming, *attestation.Status)
	assert.Equal(t, attestation.TrustVector.Executables, ear.ApprovedRuntimeClaim)
}
