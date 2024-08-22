// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

const SchemeName = "TPM_ENACTTRUST"

var (
	EndorsementMediaTypes = []string{
		`application/corim-unsigned+cbor; profile="http://enacttrust.com/veraison/1.0.0"`,
	}

	EvidenceMediaTypes = []string{
		"application/vnd.enacttrust.tpm-evidence",
	}
)
