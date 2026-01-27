// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

// NOTE: this file is generated. DO NOT EDIT

import _ "embed"

var (
	//go:embed test/corim/corim-sevsnp-bad-refval-key.cbor
	corimSevsnpBadRefvalKey []byte

	//go:embed test/corim/corim-sevsnp-bad-refval-no-key.cbor
	corimSevsnpBadRefvalNoKey []byte

	//go:embed test/corim/corim-sevsnp-bad-ta-no-model.cbor
	corimSevsnpBadTaNoModel []byte

	//go:embed test/corim/corim-sevsnp-bad-ta-no-vendor.cbor
	corimSevsnpBadTaNoVendor []byte

	//go:embed test/corim/corim-sevsnp-valid.cbor
	corimSevsnpValid []byte
)
