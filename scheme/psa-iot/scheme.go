// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

const (
	// SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
	SchemeName = "ARM_PSA_IOT"

	// Plugin name constants following the format: veraison/<scheme>/<handler-type>
	EvidenceHandlerName    = "veraison/psa-iot/evidence"
	EndorsementHandlerName = "veraison/psa-iot/endorsement"
	StoreHandlerName       = "veraison/psa-iot/store"
)

var EndorsementMediaTypes = []string{
	// Unsigned CoRIM profile
	`application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"`,
	// Signed CoRIM profile
	`application/rim+cose; profile="http://arm.com/psa/iot/1"`,
}

var EvidenceMediaTypes = []string{
	"application/psa-attestation-token",
	`application/eat-cwt; profile="http://arm.com/psa/2.0.0"`,
	`application/eat+cwt; eat_profile="tag:psacertified.org,2023:psa#tfm"`,
	`application/eat+cwt; eat_profile="tag:psacertified.org,2019:psa#legacy"`,
}
