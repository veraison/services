// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import _ "embed"

var (
  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyGood.cbor
  unsignedCorimComidParsecTpmKeyGood []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyNoClass.cbor
  unsignedCorimComidParsecTpmKeyNoClass []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyNoClassId.cbor
  unsignedCorimComidParsecTpmKeyNoClassId []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyNoInstance.cbor
  unsignedCorimComidParsecTpmKeyNoInstance []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyUnknownClassIdType.cbor
  unsignedCorimComidParsecTpmKeyUnknownClassIdType []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyUnknownInstanceType.cbor
  unsignedCorimComidParsecTpmKeyUnknownInstanceType []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmKeyManyKeys.cbor
  unsignedCorimComidParsecTpmKeyManyKeys []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmPcrsGood.cbor
  unsignedCorimComidParsecTpmPcrsGood []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmPcrsNoClass.cbor
  unsignedCorimComidParsecTpmPcrsNoClass []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmPcrsNoPCR.cbor
  unsignedCorimComidParsecTpmPcrsNoPCR []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmPcrsUnknownPCRType.cbor
  unsignedCorimComidParsecTpmPcrsUnknownPCRType []byte

  //go:embed test/corim/unsignedCorimMiniComidParsecTpmPcrsNoDigests.cbor
  unsignedCorimComidParsecTpmPcrsNoDigests []byte
)
