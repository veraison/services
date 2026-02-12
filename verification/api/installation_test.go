// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInstallationInfo_NoMetadata(t *testing.T) {
	// When no metadata file exists, should return nil without error
	info, err := GetInstallationInfoFromPaths([]string{"/nonexistent/path.json"})
	require.NoError(t, err)
	assert.Nil(t, info)
}

func TestGetInstallationInfo_ValidMetadata(t *testing.T) {
	// Create temporary metadata file
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "installation.json")

	expectedInfo := &InstallationInfo{
		Version:          "1.0.0",
		DeploymentMethod: "deb",
		InstallTime:      "2025-09-23T10:00:00Z",
		AttestationPath:  "/usr/share/doc/veraison/attestation.json",
		ArtifactDigest:   "sha256:1234567890abcdef",
		Metadata: map[string]string{
			"package": "veraison-services",
		},
	}

	err := WriteInstallationMetadata(expectedInfo, metadataPath)
	require.NoError(t, err)

	// Read it back
	info, err := GetInstallationInfoFromPaths([]string{metadataPath})
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, expectedInfo.Version, info.Version)
	assert.Equal(t, expectedInfo.DeploymentMethod, info.DeploymentMethod)
	assert.Equal(t, expectedInfo.InstallTime, info.InstallTime)
	assert.Equal(t, expectedInfo.AttestationPath, info.AttestationPath)
	assert.Equal(t, expectedInfo.ArtifactDigest, info.ArtifactDigest)
	assert.Equal(t, expectedInfo.Metadata, info.Metadata)
}

func TestGetInstallationInfo_RelativeAttestationPath(t *testing.T) {
	// Test that relative attestation paths are resolved relative to metadata file
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "installation.json")

	info := &InstallationInfo{
		Version:          "1.0.0",
		DeploymentMethod: "native",
		AttestationPath:  "attestations/artifact.json", // Relative path
	}

	err := WriteInstallationMetadata(info, metadataPath)
	require.NoError(t, err)

	// Read it back
	readInfo, err := GetInstallationInfoFromPaths([]string{metadataPath})
	require.NoError(t, err)
	require.NotNil(t, readInfo)

	// Should be resolved to absolute path
	expectedPath := filepath.Join(tmpDir, "attestations/artifact.json")
	assert.Equal(t, expectedPath, readInfo.AttestationPath)
}

func TestGetInstallationInfo_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "installation.json")

	err := os.WriteFile(metadataPath, []byte("invalid json{"), 0644)
	require.NoError(t, err)

	// Should return error
	info, err := GetInstallationInfoFromPaths([]string{metadataPath})
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "invalid metadata JSON")
}

func TestGetInstallationInfo_MultiplePathsFallback(t *testing.T) {
	// Create metadata in second path
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "installation.json")

	expectedInfo := &InstallationInfo{
		Version:          "2.0.0",
		DeploymentMethod: "docker",
	}

	err := WriteInstallationMetadata(expectedInfo, metadataPath)
	require.NoError(t, err)

	// First path doesn't exist, should fall back to second
	paths := []string{
		"/nonexistent/first.json",
		metadataPath,
		"/nonexistent/third.json",
	}

	info, err := GetInstallationInfoFromPaths(paths)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "2.0.0", info.Version)
	assert.Equal(t, "docker", info.DeploymentMethod)
}

func TestInstallationInfoJSON(t *testing.T) {
	info := InstallationInfo{
		Version:          "1.0.0",
		DeploymentMethod: "deb",
		InstallTime:      "2025-09-23T10:00:00Z",
		AttestationPath:  "/usr/share/doc/veraison/attestation.json",
		ArtifactDigest:   "sha256:1234567890abcdef",
		Metadata: map[string]string{
			"package":      "veraison-services",
			"architecture": "amd64",
		},
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded InstallationInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info, decoded)
}

func TestWriteInstallationMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	metadataPath := filepath.Join(tmpDir, "subdir", "installation.json")

	info := &InstallationInfo{
		Version:          "1.5.0",
		DeploymentMethod: "rpm",
		InstallTime:      "2025-10-01T12:00:00Z",
	}

	// Should create parent directories
	err := WriteInstallationMetadata(info, metadataPath)
	require.NoError(t, err)

	// Verify file exists and can be read
	_, err = os.Stat(metadataPath)
	require.NoError(t, err)

	// Verify content
	readInfo, err := readMetadataFile(metadataPath)
	require.NoError(t, err)
	assert.Equal(t, info.Version, readInfo.Version)
	assert.Equal(t, info.DeploymentMethod, readInfo.DeploymentMethod)
}