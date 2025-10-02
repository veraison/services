// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"mime"

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
	return EndorsementHandlerName
}

func (o EndorsementHandler) GetAttestationScheme() string {
	return SchemeName
}

func (o EndorsementHandler) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

func (o EndorsementHandler) Decode(data []byte, mediaType string, caCertPool []byte) (*handler.EndorsementHandlerResponse, error) {
	extractor := &CorimExtractor{}

	if mediaType != "" {
		mt, _, err := mime.ParseMediaType(mediaType)
		if err != nil {
			return nil, err
		}

		// Use signed decoder for signed CoRIM
		if mt == "application/rim+cose" {
			return common.SignedCorimDecoder(data, extractor, caCertPool)
		}
	}

	// Default to unsigned CoRIM decoder
	return common.UnsignedCorimDecoder(data, extractor)
}
