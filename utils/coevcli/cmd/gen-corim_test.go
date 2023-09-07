// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RootCmd_unknown_argument(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"--unknown-argument=val"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.EqualError(t, err, "unknown flag: --unknown-argument")
}

func Test_RootCmd_no_key_file(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no key supplied")
}

func Test_RootCmd_no_evidence_file(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no evidence file supplied")
}

func Test_RootCmd_no_attestation_scheme(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no attestation scheme supplied")
}

func Test_RootCmd_invalid_attestation_scheme(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=invalid-scheme",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "unsupported attestation scheme")
}

func Test_RootCmd_cocli_psa_runs(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
	os.Remove("psa-endorsements.cbor")
}

func Test_RootCmd_cocli_cca_runs(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/cca-evidence.cbor",
		"--attest-scheme=cca",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
	os.Remove("cca-endorsements.cbor")
}

func Test_RootCmd_with_output(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--corim-file=../data/corims/test-target.cbor",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	os.Remove("../data/corims/test-target.cbor")

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, "../data/corims/test-target.cbor")
	os.Remove("../data/corims/test-target.cbor")
}

func Test_RootCmd_with_wrong_key(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/ec256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_wrong_scheme(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/cca-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_evidence(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/bad-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_output_path(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/corims/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates",
		"--corim-file=../data/",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_no_template_dir(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "no template directory supplied")
}

func Test_RootCmd_with_bad_template_dir_path(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/not-exist",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "template directory does not exist")
}

func Test_RootCmd_with_missing_comid_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates/error-templates/just-corim",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "file `comid-template.json` is missing from template directory")
}

func Test_RootCmd_with_missing_corim_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates/error-templates/just-comid",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "file `corim-template.json` is missing from template directory")
}

func Test_RootCmd_with_bad_comid_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates/error-templates/bad-comid",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_corim_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"--key-file=../data/keys/es256.json",
		"--evidence-file=../data/templates/psa-evidence.cbor",
		"--attest-scheme=psa",
		"--template-dir=../data/templates/error-templates/bad-corim",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_PubKeyFromJWK_with_bad_key(t *testing.T) {
	_, err := PubKeyFromJWK(nil)
	assert.Error(t, err)
}

func Test_convertJwkToPEM_with_bad_path(t *testing.T) {
	_, err := convertJwkToPEM("")
	assert.Error(t, err)
}

func Test_convertJwkToPEM_with_pub_key(t *testing.T) {
	_, err := convertJwkToPEM("../data/keys/ec256.json")
	assert.Error(t, err)
}
func Test_convertJwkToPEM_with_bad_file(t *testing.T) {
	_, err := convertJwkToPEM("../data/templates/comid-claims-template.json")
	assert.Error(t, err)
}
