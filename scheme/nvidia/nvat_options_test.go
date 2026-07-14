// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"os"
	"testing"
)

func TestDefaultNvatOptions(t *testing.T) {
	opts := DefaultNvatOptions()

	if opts.VerifierMode != "remote" {
		t.Fatalf("expected default verifier mode 'remote', got %q", opts.VerifierMode)
	}

	if opts.CaTrustFile != "" {
		if _, err := os.Stat(opts.CaTrustFile); err != nil {
			t.Fatalf("default CA trust file %q is not accessible: %v", opts.CaTrustFile, err)
		}
	}
}

func TestNvatOptionsValidSuccess(t *testing.T) {
	opts := NvatOptions{
		VerifierMode: "remote",
		ServiceToken: "token",
	}

	if err := opts.Valid(); err != nil {
		t.Fatalf("expected options to be valid, got error: %v", err)
	}
}

func TestNvatOptionsValidLocalWithoutServiceToken(t *testing.T) {
	opts := NvatOptions{
		VerifierMode: "local",
	}

	if err := opts.Valid(); err != nil {
		t.Fatalf("expected local options without a service token to be valid, got error: %v", err)
	}
}

func TestMissingRemoteServiceToken(t *testing.T) {
	testCases := []struct {
		name string
		opts NvatOptions
		want bool
	}{
		{
			name: "remote without token",
			opts: NvatOptions{VerifierMode: "remote"},
			want: true,
		},
		{
			name: "remote with token",
			opts: NvatOptions{VerifierMode: "remote", ServiceToken: "token"},
			want: false,
		},
		{
			name: "local without token",
			opts: NvatOptions{VerifierMode: "local"},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := missingRemoteServiceToken(tc.opts); got != tc.want {
				t.Fatalf("missingRemoteServiceToken() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestNvatOptionsValidFailures(t *testing.T) {
	opts := NvatOptions{
		ServiceToken: "token",
	}
	if err := opts.Valid(); err == nil {
		t.Fatal("expected error for missing verifier mode")
	}

	opts = NvatOptions{
		VerifierMode: "invalid",
		ServiceToken: "token",
	}
	if err := opts.Valid(); err == nil {
		t.Fatal("expected error for unsupported verifier mode")
	}

	opts = NvatOptions{
		VerifierMode: "remote",
	}
	if err := opts.Valid(); err == nil {
		t.Fatal("expected error for missing service token")
	}
}
