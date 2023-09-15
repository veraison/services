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

func Test_RootCmd_with_two_args(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 3 arg(s), received 2")
}

func Test_RootCmd_invalid_attestation_scheme(t *testing.T) {
	cmd := NewRootCmd()

	args := []string{"invalid-scheme",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "unsupported attestation scheme invalid-scheme, only psa and cca are supported")
}

func Test_RootCmd_psa_runs(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
	os.Remove("psa-endorsements.cbor")
}

func Test_RootCmd_cca_runs(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"cca",
		"../data/corims/cca-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.NoError(t, err)
	os.Remove("cca-endorsements.cbor")
}

func Test_RootCmd_with_output(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
		"--corim-file=../data/corims/test-target.cbor",
	}
	cmd.SetArgs((args))

	os.Remove("../data/corims/test-target.cbor")

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, "../data/corims/test-target.cbor")
	os.Remove("../data/corims/test-target.cbor")
}

func Test_RootCmd_Execute(t *testing.T) {

	*genCorimTemplateDir = "../data/templates/psa"
	*genCorimCorimFile = ""

	os.Args = []string{"gen-corim", "psa", "../data/corims/psa-evidence.cbor", "../data/keys/es256.json"}

	Execute()
	os.Remove("psa-endorsements.cbor")
}

func Test_RootCmd_with_wrong_key(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/ec256.json",
		"--template-dir=../data/templates/psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_wrong_scheme(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/cca-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/cca",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_evidence(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/bad-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_output_path(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa",
		"--corim-file=../data/",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_no_template_dir(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "template directory does not exist")
}

func Test_RootCmd_with_bad_template_dir_path(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/not-exist",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "template directory does not exist")
}

func Test_RootCmd_with_missing_comid_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa/error-templates/just-corim",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "file `comid-template.json` is missing from template directory")
}

func Test_RootCmd_with_missing_corim_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa/error-templates/just-comid",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.EqualError(t, err, "file `corim-template.json` is missing from template directory")
}

func Test_RootCmd_with_bad_comid_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa/error-templates/bad-comid",
	}
	cmd.SetArgs((args))

	err := cmd.Execute()
	assert.Error(t, err)
}

func Test_RootCmd_with_bad_corim_template(t *testing.T) {

	cmd := NewRootCmd()

	args := []string{"psa",
		"../data/corims/psa-evidence.cbor",
		"../data/keys/es256.json",
		"--template-dir=../data/templates/psa/error-templates/bad-corim",
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
