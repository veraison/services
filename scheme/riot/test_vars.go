// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package riot

// NOTE: this file is generated. DO NOT EDIT

import _ "embed"

var (
	//go:embed test/corim/corim-riot-bad-no-vendor.cbor
	corimRiotBadNoVendor []byte

	//go:embed test/corim/corim-riot-bad-refvals.cbor
	corimRiotBadRefvals []byte

	//go:embed test/corim/corim-riot-bad-wrong-key-type.cbor
	corimRiotBadWrongKeyType []byte

	//go:embed test/corim/corim-riot-bad-wrong-vendor.cbor
	corimRiotBadWrongVendor []byte

	//go:embed test/corim/corim-riot-valid.cbor
	corimRiotValid []byte
)
