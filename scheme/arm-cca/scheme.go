// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

const SchemeName = "ARM_CCA"

var (
	EndorsementMediaTypes = []string{
		`application/corim-unsigned+cbor; profile="http://arm.com/cca/ssd/1"`,
		`application/corim-unsigned+cbor; profile="http://arm.com/cca/realm/1"`,
	}

	EvidenceMediaTypes = []string{
		`application/eat-collection; profile="http://arm.com/CCA-SSD/1.0.0"`,
	}
)
