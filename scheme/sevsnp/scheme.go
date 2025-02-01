// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

const (
	SchemeName             = "SEVSNP"
	EndorsementMediaTypeRV = `application/corim-unsigned+cbor; profile="tag:amd.com,2024:snp-corim-profile"`
	// ToDo: check media type for AMD ARK
	EndorsementMediaTypeTA = `application/corim-unsigned+cbor; profile="https://amd.com/ark"`
	EvidenceMediaType      = "application/vnd.veraison.tsm-report+cbor"
)

var (
	EndorsementMediaTypes = []string{
		EndorsementMediaTypeRV,
		EndorsementMediaTypeTA,
	}

	EvidenceMediaTypes = []string{
		EvidenceMediaType,
	}
)

const (
	mKeyMeasurement = 641
)
