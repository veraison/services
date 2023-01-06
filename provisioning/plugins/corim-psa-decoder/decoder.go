// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/veraison/services/provisioning/decoder"
	plugin_common "github.com/veraison/services/provisioning/plugins/common"
)

const (
	SupportedPSAMediaType = "application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1"
	SupportedCCAMediaType = "application/corim-unsigned+cbor; profile=http://arm.com/cca/ssd/1"
	PluginName            = "unsigned-corim (PSA & CCA platform profiles)"
)

type Decoder struct{}

func (o Decoder) Init(params decoder.Params) error {
	return nil // no-op
}

func (o Decoder) Close() error {
	return nil // no-op
}

func (o Decoder) GetName() string {
	return PluginName
}

func (o Decoder) GetSupportedMediaTypes() []string {
	return []string{
		SupportedPSAMediaType,
		SupportedCCAMediaType,
	}
}

func (o Decoder) Decode(data []byte) (*decoder.EndorsementDecoderResponse, error) {
	return plugin_common.UnsignedCorimDecoder(data, &Extractor{})
}
