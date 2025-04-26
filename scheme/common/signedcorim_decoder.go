// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"

	"github.com/veraison/corim/corim"
	"github.com/veraison/services/handler"
)

// calculateCertThumbprint computes the SHA-256 thumbprint of an X.509 certificate
func calculateCertThumbprint(cert *x509.Certificate) (string, error) {
	if cert == nil {
		return "", fmt.Errorf("certificate is nil")
	}

	thumbprint := sha256.Sum256(cert.Raw)
	hexStr := hex.EncodeToString(thumbprint[:])

	return hexStr, nil
}

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

	rootCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("could not load system certificates: %w", err)
	}

	// CA certs must be provided for verification
	if len(caCertPoolPEM) == 0 {
		return nil, fmt.Errorf("no CA certificates available for verification")
	}

	// Add the CA certificates from the provided PEM data
	if !rootCertPool.AppendCertsFromPEM(caCertPoolPEM) {
		return nil, fmt.Errorf("failed to parse and append CA certificates from the cert pool")
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

	_, err = sc.SigningCert.Verify(verifyOpts)
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

	// Calculate the signing certificate thumbprint
	thumbprint, err := calculateCertThumbprint(sc.SigningCert)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate certificate thumbprint: %w", err)
	}

	return UnsignedCorimDecoder(unsignedCorimCBOR, xtr, thumbprint)
}
