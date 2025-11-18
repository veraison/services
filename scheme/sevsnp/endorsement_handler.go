// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"errors"
	"mime"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common"
)

// EndorsementHandler implements the IEndorsementHandler interface for SEVSNP scheme
type EndorsementHandler struct{}

// Init initializes the endorsement handler instance. no-op for SEVSNP
func (o EndorsementHandler) Init(params handler.EndorsementHandlerParams) error {
	return nil // no-op
}

// Close closes the endorsement handler instance. no-op for SEVSNP
func (o EndorsementHandler) Close() error {
	return nil // no-op
}

// GetName returns the name of the endorsement handler
func (o EndorsementHandler) GetName() string {
	return SchemeName
}

// GetAttestationScheme returns the scheme name
func (o EndorsementHandler) GetAttestationScheme() string {
	return SchemeName
}

// GetSupportedMediaTypes returns the media types supported for SEVSNP endorsements
func (o EndorsementHandler) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

// Decode decodes the supplied endorsement as an unsigned CoRIM
func (o EndorsementHandler) Decode(data []byte, mediaType string, caCertPool []byte) (*handler.EndorsementHandlerResponse, error) {
	extractor := &Extractor{}

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

func (o EndorsementHandler) CoservRepackage(coservQuery string, resultSet []string) ([]byte, error) {
	return nil, errors.New("SEV-SNP CoservRepackage not implemented")
}
