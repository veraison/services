// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallationInfo contains information about how this Veraison instance
// was installed and its artifact attestations. This information is generated
// during deployment/installation and read from a metadata file.
type InstallationInfo struct {
	// Version of Veraison
	Version string `json:"version"`

	// DeploymentMethod describes how this instance was deployed
	// (e.g., "deb", "rpm", "docker", "native", "source", or any custom method)
	DeploymentMethod string `json:"deployment_method"`

	// Installation time (when package was installed)
	InstallTime string `json:"install_time,omitempty"`

	// Path to the artifact attestation file if available
	AttestationPath string `json:"attestation_path,omitempty"`

	// Attestation digest (sha256 of the installation artifact)
	ArtifactDigest string `json:"artifact_digest,omitempty"`

	// Additional metadata that may be specific to the deployment method
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Default paths to search for installation metadata, in order of preference
var defaultMetadataPaths = []string{
	"/etc/veraison/installation.json",           // System-wide installation
	"/usr/share/veraison/installation.json",     // Package manager installations
	"/opt/veraison/installation.json",           // Alternative installation location
	"./installation.json",                        // Local/development installations
}

// GetInstallationInfo reads installation metadata from a JSON file.
// It searches through default paths and returns the first valid metadata found.
// If no metadata file is found, returns nil with no error (optional feature).
func GetInstallationInfo() (*InstallationInfo, error) {
	return GetInstallationInfoFromPaths(defaultMetadataPaths)
}

// GetInstallationInfoFromPaths reads installation metadata from the first
// existing file in the provided paths.
func GetInstallationInfoFromPaths(paths []string) (*InstallationInfo, error) {
	for _, path := range paths {
		info, err := readMetadataFile(path)
		if err == nil {
			return info, nil
		}
		// If file doesn't exist, try next path
		if os.IsNotExist(err) {
			continue
		}
		// For other errors (permission, parse error), return the error
		return nil, fmt.Errorf("error reading metadata from %s: %w", path, err)
	}

	// No metadata file found - this is not necessarily an error
	// Return nil to indicate no installation info available
	return nil, nil
}

// readMetadataFile reads and parses installation metadata from a JSON file
func readMetadataFile(path string) (*InstallationInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var info InstallationInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("invalid metadata JSON: %w", err)
	}

	// Resolve relative attestation path if present
	if info.AttestationPath != "" && !filepath.IsAbs(info.AttestationPath) {
		// Make attestation path relative to metadata file location
		baseDir := filepath.Dir(path)
		info.AttestationPath = filepath.Join(baseDir, info.AttestationPath)
	}

	return &info, nil
}

// WriteInstallationMetadata writes installation metadata to a file.
// This is a helper function for use during deployment/installation.
func WriteInstallationMetadata(info *InstallationInfo, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata to %s: %w", path, err)
	}

	return nil
}