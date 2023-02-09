// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_ssd_platform

const SchemeName = "CCA_SSD_PLATFORM"

var (
	EndorsementMediaTypes = []string{
		"application/corim-unsigned+cbor; profile=http://arm.com/cca/ssd/1",
	}

	EvidenceMediaTypes = []string{
		"application/eat-collection; profile=http://arm.com/CCA-SSD/1.0.0",
	}
)
