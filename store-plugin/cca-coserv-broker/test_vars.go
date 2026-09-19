// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ccacoservbroker

import _ "embed"

var (
	//go:embed test/coserv/coserv_ta_result.cbor
	CoSERVTrustAnchorResponse []byte

	//go:embed test/coserv/coserv_rv_result.cbor
	CoSERVReferenceValueResponse []byte
)
