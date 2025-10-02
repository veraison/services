// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package riot

const (
	// SchemeName follows the format: <VENDOR>_<TECHNOLOGY>_<VARIANT>
	// RIoT is standardized by TCG, so we use RIOT as the scheme name
	SchemeName = "RIOT"

	// Plugin name constants following the format: veraison/<scheme>/<handler-type>
	EvidenceHandlerName = "veraison/riot/evidence"
	StoreHandlerName    = "veraison/riot/store"
	// Note: RIoT doesn't have an endorsement handler
)

var EvidenceMediaTypes = []string{
	"application/pem-certificate-chain",
}
