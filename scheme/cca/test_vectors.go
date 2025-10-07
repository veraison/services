// Copyright 2024 Contributors to the Veraison project.
// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import _ "embed"

var (
	//go:embed test/corim/unsignedCorimCcaComidCcaRefValOne.cbor
	unsignedCorimCcaComidCcaRefValOne []byte

	//go:embed test/corim/unsignedCorimCcaComidCcaRefValFour.cbor
	unsignedCorimCcaComidCcaRefValFour []byte

	//go:embed test/corim/unsignedCorimCcaNoProfileComidCcaRefValOne.cbor
	unsignedCorimCcaNoProfileComidCcaRefValOne []byte

	//go:embed test/corim/unsignedCorimCcaNoProfileComidCcaRefValFour.cbor
	unsignedCorimCcaNoProfileComidCcaRefValFour []byte

	//go:embed test/corim/unsignedCorimCcaRealmComidCcaRealm.cbor
	unsignedCorimCcaRealmComidCcaRealm []byte

	//go:embed test/corim/unsignedCorimCcaRealmComidCcaRealmNoClass.cbor
	unsignedCorimCcaRealmComidCcaRealmNoClass []byte

	//go:embed test/corim/unsignedCorimCcaRealmComidCcaRealmNoInstance.cbor
	unsignedCorimCcaRealmComidCcaRealmNoInstance []byte

	//go:embed test/corim/unsignedCorimCcaRealmComidCcaRealmInvalidInstance.cbor
	unsignedCorimCcaRealmComidCcaRealmInvalidInstance []byte

	//go:embed test/corim/unsignedCorimCcaRealmComidCcaRealmInvalidClass.cbor
	unsignedCorimCcaRealmComidCcaRealmInvalidClass []byte
)
