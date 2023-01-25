// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"github.com/veraison/services/plugin"
)

func RegisterEndorsementDecoder(i IEndorsementDecoder) {
	err := plugin.RegisterImplementation("endorsement-decoder", i, EndorsementDecoderRPC)
	if err != nil {
		panic(err)
	}
}

func RegisterEvidenceDecoder(i IEvidenceDecoder) {
	err := plugin.RegisterImplementation("evidence-decoder", i, EvidenceDecoderRPC)
	if err != nil {
		panic(err)
	}
}
