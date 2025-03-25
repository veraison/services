// Copyright 2023-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

const (
	SchemeName = "PSA_IOT"
)

var EndorsementMediaTypes = []string{
	`application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"`,
}

var EvidenceMediaTypes = []string{
	"application/psa-attestation-token",
	`application/eat-cwt; profile="http://arm.com/psa/2.0.0"`,
	`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
	`application/eat+cwt; eat_profile="tag:psacertified.org,2019:psa#legacy"`,
}

var CoservMediaTypes = []string{
	`application/coserv+cbor; profile="tag:psacertified.org,2023:psa#tfm"`,
}
