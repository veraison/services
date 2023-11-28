// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm

const SchemeName = "CCA_REALM"

var (
	EndorsementMediaTypes = []string{
		"application/corim-unsigned+cbor; profile=http://arm.com/cca/realm/1",
	}

	EvidenceMediaTypes = []string{
		"application/eat-collection; profile=http://arm.com/CCA-SSD/1.0.0",
	}
)
