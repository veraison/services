// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"github.com/veraison/services/decoder"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

type EndorsementDecoder struct{}

func (o EndorsementDecoder) Init(params decoder.EndorsementDecoderParams) error {
	return nil // no-op
}

func (o EndorsementDecoder) Close() error {
	return nil // no-op
}

func (o EndorsementDecoder) GetName() string {
	return "unsigned-corim (PSA profile)"
}

func (o EndorsementDecoder) GetAttestationScheme() string {
	return SchemeName
}

func (o EndorsementDecoder) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

func (o EndorsementDecoder) Decode(data []byte) (*decoder.EndorsementDecoderResponse, error) {
	return common.UnsignedCorimDecoder(data, &arm.Extractor{Scheme: SchemeName})
}
