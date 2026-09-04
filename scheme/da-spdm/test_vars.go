// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package da_spdm

// NOTE: this file is generated. DO NOT EDIT

import _ "embed"

var (
	//go:embed test/evidence/da-spdm-good.cbor
	daSpdmGood []byte

	//go:embed test/corim/corim-da-bad-group-set.cbor
	corimDaBadGroupSet []byte

	//go:embed test/corim/corim-da-bad-instance-set.cbor
	corimDaBadInstanceSet []byte

	//go:embed test/corim/corim-da-bad-mkey-not-set.cbor
	corimDaBadMkeyNotSet []byte

	//go:embed test/corim/corim-da-bad-mkey-wrong-type.cbor
	corimDaBadMkeyWrongType []byte

	//go:embed test/corim/corim-da-bad-multiple-values.cbor
	corimDaBadMultipleValues []byte

	//go:embed test/corim/corim-da-bad-no-class.cbor
	corimDaBadNoClass []byte

	//go:embed test/corim/corim-da-bad-no-class-id.cbor
	corimDaBadNoClassId []byte

	//go:embed test/corim/corim-da-bad-no-name.cbor
	corimDaBadNoName []byte

	//go:embed test/corim/corim-da-bad-no-value.cbor
	corimDaBadNoValue []byte

	//go:embed test/corim/corim-da-ref.cbor
	corimDaRef []byte
)
