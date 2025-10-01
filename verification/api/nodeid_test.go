// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeID(t *testing.T) {
	// Set up a temporary directory for testing
	tmpDir := t.TempDir()
	os.Setenv("VERAISON_NODE_ID_DIR", tmpDir)
	defer os.Unsetenv("VERAISON_NODE_ID_DIR")

	// First call should generate and persist a node ID
	id1, err := getNodeID()
	require.NoError(t, err)
	require.Len(t, id1, nodeIDLength)

	// Second call should read the same persisted ID
	id2, err := getNodeID()
	require.NoError(t, err)
	assert.Equal(t, id1, id2)

	// Verify file contents
	data, err := os.ReadFile(filepath.Join(tmpDir, nodeIDFileName))
	require.NoError(t, err)
	decoded, err := hex.DecodeString(string(data))
	require.NoError(t, err)
	assert.Equal(t, id1, decoded)
}

func TestGenerateRandomNodeID(t *testing.T) {
	id, err := generateRandomNodeID()
	require.NoError(t, err)
	require.Len(t, id, nodeIDLength)
	// Check multicast bit is set
	assert.True(t, id[0]&0x01 == 0x01)

	// Generate another to ensure they're different
	id2, err := generateRandomNodeID()
	require.NoError(t, err)
	assert.NotEqual(t, id, id2)
}

func TestGetMACBasedID(t *testing.T) {
	// This test might be skipped if no suitable interface is found
	id, err := getMACBasedID()
	if err != nil {
		t.Skip("No suitable network interface found for testing")
	}
	require.Len(t, id, nodeIDLength)
}

func TestGetNodeIDDirDefault(t *testing.T) {
	os.Unsetenv("VERAISON_NODE_ID_DIR")
	assert.Equal(t, "/var/lib/veraison", getNodeIDDir())
}

func TestGetNodeIDDirCustom(t *testing.T) {
	os.Setenv("VERAISON_NODE_ID_DIR", "/custom/path")
	defer os.Unsetenv("VERAISON_NODE_ID_DIR")
	assert.Equal(t, "/custom/path", getNodeIDDir())
}