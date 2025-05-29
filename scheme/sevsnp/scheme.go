// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

const (
	SchemeName             = "SEVSNP"
	EndorsementMediaTypeRV = `application/corim-unsigned+cbor; profile="tag:amd.com,2024:snp-corim-profile"`
	// ToDo: check media type for AMD ARK
	EndorsementMediaTypeTA   = `application/corim-unsigned+cbor; profile="https://amd.com/ark"`
	EvidenceMediaTypeTSMCbor = "application/vnd.veraison.tsm-report+cbor"
	EvidenceMediaTypeTSMJson = "application/vnd.veraison.configfs-tsm+json"
	EvidenceMediaTypeRATSd   = `application/eat+cwt; eat_profile="tag:github.com,2025:veraison/ratsd/cmw"`
)

var (
	EndorsementMediaTypes = []string{
		EndorsementMediaTypeRV,
		EndorsementMediaTypeTA,
	}

	EvidenceMediaTypes = []string{
		EvidenceMediaTypeTSMCbor,
		EvidenceMediaTypeTSMJson,
		EvidenceMediaTypeRATSd,
	}
)

const (
	mKeyReportData  = 640
	mKeyMeasurement = 641
	mKeyReportID    = 645
	mKeyReportedTcb = 647
)
