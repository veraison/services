// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"crypto/x509"
	"fmt"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common"
)

type EndorsementHandler struct {
	caPool *x509.CertPool
}

func (o EndorsementHandler) Init(params handler.EndorsementHandlerParams) error {
	// Extract CA certificates from params
	if caCerts, ok := params["ca_certs"].([]*x509.Certificate); ok {
		o.caPool = x509.NewCertPool()
		for _, cert := range caCerts {
			o.caPool.AddCert(cert)
		}
	}
	return nil
}

func (o EndorsementHandler) Close() error {
	return nil // no-op
}

func (o EndorsementHandler) GetName() string {
	return "corim (Parsec TPM profile)"
}

func (o EndorsementHandler) GetAttestationScheme() string {
	return SchemeName
}

func (o EndorsementHandler) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

func (o EndorsementHandler) Decode(data []byte, mediaType string) (*handler.EndorsementHandlerResponse, error) {
	switch mediaType {
	case "application/rim+cbor":
		return common.UnsignedCorimDecoder(data, &CorimExtractor{})
	case "application/rim+cose":
		if o.caPool == nil {
			return nil, fmt.Errorf("CA certificate pool not initialized")
		}
		return common.SignedCorimDecoder(data, &CorimExtractor{}, o.caPool)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
}
