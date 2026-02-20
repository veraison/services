// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

// NOTE: this file is generated. DO NOT EDIT

import _ "embed"

var (
	//go:embed test/corim/corim-psa-bad-class.cbor
	corimPsaBadClass []byte

	//go:embed test/corim/corim-psa-bad-instance.cbor
	corimPsaBadInstance []byte

	//go:embed test/corim/corim-psa-bad-refval-instance.cbor
	corimPsaBadRefvalInstance []byte

	//go:embed test/corim/corim-psa-bad-refval-mkey.cbor
	corimPsaBadRefvalMkey []byte

	//go:embed test/corim/corim-psa-bad-refval-mval.cbor
	corimPsaBadRefvalMval []byte

	//go:embed test/corim/corim-psa-bad-ta-cert.cbor
	corimPsaBadTaCert []byte

	//go:embed test/corim/corim-psa-bad-ta-no-instance.cbor
	corimPsaBadTaNoInstance []byte

	//go:embed test/corim/corim-psa-valid.cbor
	corimPsaValid []byte
)
