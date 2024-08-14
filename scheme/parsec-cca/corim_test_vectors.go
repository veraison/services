// Copyright 2023-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package parsec_cca

import _ "embed"

var (
	//go:embed test/corim/unsignedCorimParsecCcaComidParsecCcaRefValOne.cbor
	unsignedCorimComidParsecCcaRefValOne []byte

	//go:embed test/corim/unsignedCorimParsecCcaComidParsecCcaMultRefVal.cbor
	unsignedCorimComidParsecCcaMultRefVal []byte
)
