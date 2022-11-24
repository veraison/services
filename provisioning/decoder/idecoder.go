// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"github.com/veraison/services/plugin"
)

type Params map[string]interface{}

type IDecoder interface {
	plugin.IPluggable
	Init(params Params) error
	Close() error
	Decode([]byte) (*EndorsementDecoderResponse, error)
}

func RegisterImplementation(i IDecoder) {
	plugin.RegisterImplementation("decoder", i, DecoderRPC)
}
