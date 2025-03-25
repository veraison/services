// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"errors"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common"
)

type EndorsementHandler struct{}

func (o EndorsementHandler) Init(params handler.EndorsementHandlerParams) error {
	return nil // no-op
}

func (o EndorsementHandler) Close() error {
	return nil // no-op
}

func (o EndorsementHandler) GetName() string {
	return "unsigned-corim (Parsec CCA profile)"
}

func (o EndorsementHandler) GetAttestationScheme() string {
	return SchemeName
}

func (o EndorsementHandler) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

func (o EndorsementHandler) Decode(data []byte) (*handler.EndorsementHandlerResponse, error) {
	return common.UnsignedCorimDecoder(data, &ParsecCcaExtractor{})
}

func (o EndorsementHandler) CoservRepackage(coservQuery string, resultSet []string) ([]byte, error) {
	return nil, errors.New("TODO")
}
