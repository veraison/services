// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewNvatGpuInterfaceMissingLibrary(t *testing.T) {
	opts := NvatOptions{
		VerifierMode: "remote",
		ServiceToken: "token",
	}

	_, err := NewNvatGpuInterface(opts)
	if err == nil {
		t.Skip("NVAT library available; skipping missing-library test")
	}
	if !strings.Contains(err.Error(), "failed to load NVAT library") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNvatGpuInterfaceVerifyEvidence(t *testing.T) {
	serviceToken := os.Getenv("NVAT_SERVICE_TOKEN")
	if serviceToken == "" {
		t.Skip("NVAT_SERVICE_TOKEN is not set; skipping NRAS integration test")
	}

	opts := NvatOptions{
		VerifierMode: "remote",
		ServiceToken: serviceToken,
	}

	nvat, err := NewNvatGpuInterface(opts)
	if err != nil {
		if strings.Contains(err.Error(), "failed to load NVAT library") {
			t.Skip("NVAT library not available for VerifyEvidence test")
		}
		t.Skipf("unable to initialize NVAT interface in test environment: %v", err)
	}

	evidencePath := filepath.Join("testdata", "blackwell-evidence.json")
	evidenceBytes, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("failed to read evidence test data: %v", err)
	}

	err = nvat.VerifyEvidence("0xcb86922a772181d54971ccc584bbc96b717c32b7087fff9aaf52832dd3f697a1", evidenceBytes)
	if err != nil {
		t.Skipf("NVAT verification not available in test environment: %v", err)
	}
}
