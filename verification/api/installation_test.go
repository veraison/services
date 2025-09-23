// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInstallationInfo(t *testing.T) {
	info, err := GetInstallationInfo()
	require.NoError(t, err)
	require.NotNil(t, info)

	// Installation type should be one of the supported types
	assert.Contains(t, []string{"deb", "rpm", "container", "native"}, info.InstallType)

	// Version should not be empty
	assert.NotEmpty(t, info.Version)
}

func TestIsContainer(t *testing.T) {
	// Note: This test may need to be skipped depending on the environment
	isContainer := isContainer()
	t.Logf("Running in container: %v", isContainer)
}

func TestInstallationInfoJSON(t *testing.T) {
	info := InstallationInfo{
		Version:          "1.0.0",
		InstallType:      "deb",
		InstallTime:      "2025-09-23T10:00:00Z",
		AttestationPath:  "/usr/share/doc/veraison/attestation.json",
		ArtifactDigest:   "sha256:1234567890abcdef",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded InstallationInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info, decoded)
}