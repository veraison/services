// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

const (
	SchemeName         = "PARSEC_TPM"
	EndorsementProfile = `"tag:github.com/parallaxsecond,2023-03-03:tpm"`
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
