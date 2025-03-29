// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/veraison/corim/corim"
	"github.com/veraison/services/handler"
)

// SignedCorimDecoder decodes a signed CoRIM message, verifies its signature,
// and extracts endorsements using the provided extractor.
func SignedCorimDecoder(
	data []byte,
	xtr IExtractor,
	caPool *x509.CertPool,
) (*handler.EndorsementHandlerResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	var sc corim.SignedCorim
	if err := sc.FromCOSE(data); err != nil {
		return nil, fmt.Errorf("COSE decoding failed: %w", err)
	}

	// Extract and verify certificate chain
	if sc.SigningCert == nil {
		return nil, errors.New("no signing certificate found in CoRIM")
	}

	// Build certificate chain
	chain := []*x509.Certificate{sc.SigningCert}
	chain = append(chain, sc.IntermediateCerts...)

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:         caPool,
		Intermediates: x509.NewCertPool(),
		CurrentTime:   sc.SigningCert.NotBefore,
	}

	// Add intermediate certificates to the pool
	for _, cert := range sc.IntermediateCerts {
		opts.Intermediates.AddCert(cert)
	}

	if _, err := chain[0].Verify(opts); err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Verify the signature using the public key from the signing certificate
	if err := sc.Verify(sc.SigningCert.PublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Now that we've verified the signature, we can process the unsigned CoRIM
	// inside the signed CoRIM
	unsignedCorimCBOR, err := sc.UnsignedCorim.ToCBOR()
	if err != nil {
		return nil, fmt.Errorf("failed to encode unsigned CoRIM: %w", err)
	}

	return UnsignedCorimDecoder(unsignedCorimCBOR, xtr)
}
