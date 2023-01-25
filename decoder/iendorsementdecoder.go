// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"github.com/veraison/services/plugin"
)

type EndorsementDecoderParams map[string]interface{}

type IEndorsementDecoder interface {
	plugin.IPluggable
	Init(params EndorsementDecoderParams) error
	Close() error
	Decode([]byte) (*EndorsementDecoderResponse, error)
}
