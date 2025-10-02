// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

const (
	// SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
	SchemeName = "TPM_ENACTTRUST"

	// Plugin name constants following the format: veraison/<scheme>/<handler-type>
	EvidenceHandlerName    = "veraison/tpm-enacttrust/evidence"
	EndorsementHandlerName = "veraison/tpm-enacttrust/endorsement"
	StoreHandlerName       = "veraison/tpm-enacttrust/store"
)

var (
	EndorsementMediaTypes = []string{
		// Unsigned CoRIM profiles
		`application/corim-unsigned+cbor; profile="http://enacttrust.com/veraison/1.0.0"`,
		// Signed CoRIM profiles
		`application/rim+cose; profile="http://enacttrust.com/veraison/1.0.0"`,
	}

	EvidenceMediaTypes = []string{
		"application/vnd.enacttrust.tpm-evidence",
	}
)
