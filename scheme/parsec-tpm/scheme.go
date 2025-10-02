// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

const (
	// SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
	SchemeName         = "PARSEC_TPM"
	EndorsementProfile = `"tag:github.com/parallaxsecond,2023-03-03:tpm"`

	// Plugin name constants following the format: veraison/<scheme>/<handler-type>
	EvidenceHandlerName    = "veraison/parsec-tpm/evidence"
	EndorsementHandlerName = "veraison/parsec-tpm/endorsement"
	StoreHandlerName       = "veraison/parsec-tpm/store"
)

var EndorsementMediaTypes = []string{
	// Unsigned CoRIM profiles
	`application/corim-unsigned+cbor; profile=` + EndorsementProfile,
	// Signed CoRIM profiles
	`application/rim+cose; profile=` + EndorsementProfile,
}

var EvidenceMediaTypes = []string{
	"application/vnd.parallaxsecond.key-attestation.tpm",
}
