// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenCmd_unknown_argument(t *testing.T) {
	cmd := NewGenCmd()

	args := []string{"--unknown-argument=val"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.EqualError(t, err, "unknown flag: --unknown-argument")
}

func Test_GenCmd_no_key_file(t *testing.T) {
	cmd := NewGenCmd()

	args := []string{"--attest-scheme=psa",
		"--evidence-file=evidence.cbor",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no key supplied")
}

func Test_GenCmd_no_evidence_file(t *testing.T) {
	cmd := NewGenCmd()

	args := []string{"--attest-scheme=psa",
		"--key-file=key.cbor",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no evidence file supplied")
}

func Test_GenCmd_no_attestation_scheme(t *testing.T) {
	cmd := NewGenCmd()

	args := []string{"--key-file=key.cbor",
		"--evidence-file=evidence.cbor",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no attestation scheme supplied")
}

func Test_GenCmd_invalid_attestation_scheme(t *testing.T) {
	cmd := NewGenCmd()

	args := []string{"--key-file=key.cbor",
		"--evidence-file=evidence.cbor",
		"--attest-scheme=invalid-scheme",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "unsupported attestation scheme")
}

func Test_GenCmd_cocli_psa_runs(t *testing.T) {

	cmd := NewGenCmd()

	args := []string{"--key-file=../data/es256.json",
		"--evidence-file=../data/psa-evidence.cbor",
		"--attest-scheme=psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
}

func Test_GenCmd_cocli_cca_runs(t *testing.T) {

	cmd := NewGenCmd()

	args := []string{"--key-file=../data/es256.json",
		"--evidence-file=../data/cca-evidence.cbor",
		"--attest-scheme=cca",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
}

func Test_GenCmd_with_output(t *testing.T) {

	cmd := NewGenCmd()

	args := []string{"--key-file=../data/es256.json",
		"--evidence-file=../data/cca-evidence.cbor",
		"--attest-scheme=cca",
		"--corim-file=../data/test-target.cbor",
	}
	cmd.SetArgs((args))

	os.Remove("../data/test-target.cbor")

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, "../data/test-target.cbor")
}
