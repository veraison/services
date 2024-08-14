// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import _ "embed"

var (
	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaIakPubOne.cbor
	unsignedCorimComidPsaIakPubOne []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaIakPubTwo.cbor
	unsignedCorimComidPsaIakPubTwo []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValOne.cbor
	unsignedCorimComidPsaRefValOne []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValThree.cbor
	unsignedCorimComidPsaRefValThree []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaMultIak.cbor
	unsignedCorimComidPsaMultIak []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValMultDigest.cbor
	unsignedCorimComidPsaRefValMultDigest []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValOnlyMandIDAttr.cbor
	unsignedCorimComidPsaRefValOnlyMandIDAttr []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValNoMkey.cbor
	unsignedCorimComidPsaRefValNoMkey []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaRefValNoImplID.cbor
	unsignedCorimComidPsaRefValNoImplID []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaIakPubNoUeID.cbor
	unsignedCorimComidPsaIakPubNoUeID []byte

	// nolint:unused
	//go:embed test/corim/unsignedCorimMiniComidPsaIakPubNoImplID.cbor
	unsignedCorimComidPsaIakPubNoImplID []byte
)
