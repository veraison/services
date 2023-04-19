// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

const (
	SchemeName         = "PARSEC_TPM"
	EndorsementProfile = `"tag:github.com/parallaxsecond,2023-03-03:tpm"`
)

var EndorsementMediaTypes = []string{
	`application/corim-unsigned+cbor; profile=` + EndorsementProfile,
}

var EvidenceMediaTypes = []string{
	"application/vnd.parallaxsecond.key-attestation.tpm",
}
