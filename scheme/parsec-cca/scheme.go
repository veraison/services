// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

const (
	// SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
	SchemeName         = "PARSEC_CCA"
	EndorsementProfile = `"tag:github.com/parallaxsecond,2023-03-03:cca"`

	// Plugin name constants following the format: veraison/<scheme>/<handler-type>
	EvidenceHandlerName    = "veraison/parsec-cca/evidence"
	EndorsementHandlerName = "veraison/parsec-cca/endorsement"
	StoreHandlerName       = "veraison/parsec-cca/store"
)

var EndorsementMediaTypes = []string{
	// Unsigned CoRIM profile
	`application/corim-unsigned+cbor; profile=` + EndorsementProfile,
	// Signed CoRIM profile
	`application/rim+cose; profile=` + EndorsementProfile,
}

var EvidenceMediaTypes = []string{
	"application/vnd.parallaxsecond.key-attestation.cca",
}
