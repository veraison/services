// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	ratsdtoken "github.com/veraison/ratsd/ratsd-token"
	"github.com/veraison/services/plugin"
)

var (
	testImplementation *Implementation
	testNvatOptions    NvatOptions
	testInitErr        error
)

type failingGpuEvidenceVerifier struct {
	err error
}

func (v failingGpuEvidenceVerifier) VerifyEvidence(_ string, _ []byte) error {
	return v.err
}

func TestMain(m *testing.M) {
	testImplementation = NewImplementation()
	testNvatOptions = nvatIntegrationTestOptions()
	if !missingRemoteServiceToken(testNvatOptions) {
		testInitErr = testImplementation.Init(
			plugin.NewParameters().
				SetString("verifier_mode", testNvatOptions.VerifierMode).
				SetString("service_token", testNvatOptions.ServiceToken).
				SetString("remote_host", testNvatOptions.RemoteHost),
		)
	}

	os.Exit(m.Run())
}

func TestImplementationAppraiseClaims(t *testing.T) {
	if missingRemoteServiceToken(testNvatOptions) {
		t.Skip("NVAT_SERVICE_TOKEN is not set; skipping NRAS integration test")
	}

	if testInitErr != nil {
		if strings.Contains(testInitErr.Error(), "failed to load NVAT library") {
			t.Skip("NVAT library not available for AppraiseClaims test")
		}
		t.Fatalf("failed to initialize NVAT interface: %v", testInitErr)
	}

	attestationReport, err := os.ReadFile(
		filepath.Join("testdata", "blackwell-evidence.json"),
	)
	require.NoError(t, err)

	nonce, err := hex.DecodeString(
		"cb86922a772181d54971ccc584bbc96b717c32b7087fff9aaf52832dd3f697a1",
	)
	require.NoError(t, err)

	claims := map[string]any{
		attestationReportClaimName: attestationReport,
		nonceClaimName:             nonce,
	}

	_, err = testImplementation.AppraiseClaims(claims, nil)
	require.NoError(t, err)
}

func TestImplementationAppraiseClaimsReturnsVerificationFailure(t *testing.T) {
	impl := NewImplementation()
	impl.nvat = failingGpuEvidenceVerifier{err: errors.New("verification failed")}

	result, err := impl.AppraiseClaims(map[string]any{
		attestationReportClaimName: []byte("attestation report"),
		nonceClaimName:             []byte("nonce"),
	}, nil)

	require.NotNil(t, result)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to verify NVIDIA GPU evidence")
}

func TestAdjustNonceMissingFunction(t *testing.T) {
	claims := ratsdtoken.Claims{
		NonceAdjustMap: map[string]uint{
			nvidiaGpuNonceAdjustMapKey: 32,
		},
	}

	adjusted, err := adjustNonce([]byte("nonce"), claims)

	require.Nil(t, adjusted)
	require.ErrorIs(t, err, ErrNonceAdjustFunctionMissing)
}
