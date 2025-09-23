// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// InstallationInfo contains information about how this Veraison instance
// was installed and its artifact attestations
type InstallationInfo struct {
	// Version of Veraison
	Version string `json:"version"`

	// Type of installation (deb, rpm, native, container)
	InstallType string `json:"install_type"`

	// Installation time (when package was installed)
	InstallTime string `json:"install_time,omitempty"`

	// Path to the artifact attestation file if available
	AttestationPath string `json:"attestation_path,omitempty"`

	// Attestation digest (sha256 of the installation artifact)
	ArtifactDigest string `json:"artifact_digest,omitempty"`
}

// GetInstallationInfo attempts to gather information about how this instance
// was installed and any available attestations
func GetInstallationInfo() (*InstallationInfo, error) {
	info := &InstallationInfo{}

	// Try to detect installation type and gather info
	if isDebPackage() {
		if err := getDebianInfo(info); err != nil {
			return nil, err
		}
	} else if isRpmPackage() {
		if err := getRpmInfo(info); err != nil {
			return nil, err
		}
	} else if isContainer() {
		if err := getContainerInfo(info); err != nil {
			return nil, err
		}
	} else {
		// Assume native installation
		if err := getNativeInfo(info); err != nil {
			return nil, err
		}
	}

	return info, nil
}

// isDebPackage checks if this is a Debian package installation
func isDebPackage() bool {
	_, err := os.Stat("/var/lib/dpkg/status")
	return err == nil
}

// isRpmPackage checks if this is an RPM package installation
func isRpmPackage() bool {
	_, err := os.Stat("/var/lib/rpm/Packages")
	return err == nil
}

// isContainer checks if running in a container
func isContainer() bool {
	_, err := os.Stat("/.dockerenv")
	if err == nil {
		return true
	}
	
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "docker") || 
	       strings.Contains(string(data), "containerd")
}

// getDebianInfo gathers installation info from dpkg
func getDebianInfo(info *InstallationInfo) error {
	info.InstallType = "deb"
	// Implementation to get deb package details
	// TODO: Get version from dpkg-query
	// TODO: Get installation time from /var/lib/dpkg/info
	// TODO: Look for attestation file in /usr/share/doc/veraison/
	return nil
}

// getRpmInfo gathers installation info from RPM
func getRpmInfo(info *InstallationInfo) error {
	info.InstallType = "rpm"
	// Implementation to get rpm package details
	// TODO: Get version from rpm -q
	// TODO: Get installation time from rpm database
	// TODO: Look for attestation file in /usr/share/doc/veraison/
	return nil
}

// getContainerInfo gathers installation info for container deployments
func getContainerInfo(info *InstallationInfo) error {
	info.InstallType = "container"
	// Implementation for container deployments
	// TODO: Get version from environment variable
	// TODO: Get container creation time
	// TODO: Look for attestation file in predefined location
	return nil
}

// getNativeInfo gathers installation info for native deployments
func getNativeInfo(info *InstallationInfo) error {
	info.InstallType = "native"
	// Implementation for native deployments
	// TODO: Get version from build info
	// TODO: Get installation time from directory timestamp
	// TODO: Look for attestation file in installation directory
	return nil
}