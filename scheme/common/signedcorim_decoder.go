// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto/x509"
	"fmt"

	"github.com/veraison/corim/corim"
	"github.com/veraison/services/handler"
)

// SignedCorimDecoder processes a signed CoRIM, verifies its signature, and then
// passes the unsigned CoRIM to the UnsignedCorimDecoder for further processing.
func SignedCorimDecoder(
	data []byte,
	xtr IExtractor,
	caCertPoolPEM []byte,
) (*handler.EndorsementHandlerResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty corim data")
	}

	// Parse the signed CoRIM which extracts certificate chain automatically through extractX5Chain
	sc := corim.NewSignedCorim()
	if err := sc.FromCOSE(data); err != nil {
		return nil, fmt.Errorf("failed to parse signed CoRIM: %w", err)
	}

	if sc.SigningCert == nil {
		return nil, fmt.Errorf("no signing certificate found in the CoRIM")
	}

	// CA certs must be provided for verification
	if len(caCertPoolPEM) == 0 {
		return nil, fmt.Errorf("no CA certificates available for verification")
	}

	// Load CA certificates from PEM into a certificate pool
	rootCertPool := x509.NewCertPool()

	// Add the CA certificates from the provided PEM data
	if !rootCertPool.AppendCertsFromPEM(caCertPoolPEM) {
		// Also try to parse as DER format
		cert, err := x509.ParseCertificate(caCertPoolPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA certificates (neither PEM nor DER format)")
		}
		rootCertPool.AddCert(cert)
	}

	// Create a separate pool for intermediate certificates
	intermediateCertPool := x509.NewCertPool()
	for _, cert := range sc.IntermediateCerts {
		intermediateCertPool.AddCert(cert)
	}

	// Verify the certificate chain with properly separated root and intermediate pools
	verifyOpts := x509.VerifyOptions{
		Roots:         rootCertPool,
		Intermediates: intermediateCertPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, err := sc.SigningCert.Verify(verifyOpts)
	if err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Verify the signature using the signing certificate's public key
	if err := sc.Verify(sc.SigningCert.PublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	unsignedCorimCBOR, err := sc.UnsignedCorim.ToCBOR()
	if err != nil {
		return nil, fmt.Errorf("failed to extract unsigned CoRIM: %w", err)
	}

	return UnsignedCorimDecoder(unsignedCorimCBOR, xtr)
}
