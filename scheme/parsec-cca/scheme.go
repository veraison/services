// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

const (
	SchemeName         = "PARSEC_CCA"
	SchemeVersion      = "1.0.0"
	EndorsementProfile = `"tag:github.com/parallaxsecond,2023-03-03:cca"`
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
