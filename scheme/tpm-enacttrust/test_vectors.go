// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import _ "embed"

var (
	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustAKOne.cbor
	unsignedCorimComidTpmEnactTrustAKOne []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustGoldenOne.cbor
	unsignedCorimComidTpmEnactTrustGoldenOne []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustAKMult.cbor
	unsignedCorimComidTpmEnactTrustAKMult []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustBadInst.cbor
	unsignedCorimComidTpmEnactTrustBadInst []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustNoInst.cbor
	unsignedCorimComidTpmEnactTrustNoInst []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustMultDigest.cbor
	unsignedCorimComidTpmEnactTrustMultDigest []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustGoldenTwo.cbor
	unsignedCorimComidTpmEnactTrustGoldenTwo []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustNoDigest.cbor
	unsignedCorimComidTpmEnactTrustNoDigest []byte

	//go:embed test/corim/unsignedCorimMiniComidTpmEnactTrustAKBadInst.cbor
	unsignedCorimComidTpmEnactTrustAKBadInst []byte
)
