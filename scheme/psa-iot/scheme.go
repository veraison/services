// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

const (
	SchemeName              = "PSA_IOT"
	CCAEndorsementMediaType = "application/corim-unsigned+cbor; profile=http://arm.com/cca/ssd/1"
)

var EndorsementMediaTypes = []string{
	"application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1",
}

var EvidenceMediaTypes = []string{
	"application/psa-attestation-token",
	"application/eat-cwt; profile=http://arm.com/psa/2.0.0",
}
