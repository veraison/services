// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

// NOTE: this file is generated. DO NOT EDIT

import _ "embed"

var (
	//go:embed test/corim/corim-enacttrust-bad-class.cbor
	corimEnacttrustBadClass []byte

	//go:embed test/corim/corim-enacttrust-bad-instance.cbor
	corimEnacttrustBadInstance []byte

	//go:embed test/corim/corim-enacttrust-bad-multiple-digests.cbor
	corimEnacttrustBadMultipleDigests []byte

	//go:embed test/corim/corim-enacttrust-bad-multiple-keys.cbor
	corimEnacttrustBadMultipleKeys []byte

	//go:embed test/corim/corim-enacttrust-bad-multiple-measurements.cbor
	corimEnacttrustBadMultipleMeasurements []byte

	//go:embed test/corim/corim-enacttrust-bad-no-digest.cbor
	corimEnacttrustBadNoDigest []byte

	//go:embed test/corim/corim-enacttrust-bad-no-instance.cbor
	corimEnacttrustBadNoInstance []byte

	//go:embed test/corim/corim-enacttrust-valid.cbor
	corimEnacttrustValid []byte
)
