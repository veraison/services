// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"google.golang.org/grpc/credentials"
)

func LoadTLSCreds(
	certPath, keyPath string,
	caPaths []string,
) (credentials.TransportCredentials, error) {
	if certPath == "" {
		return nil, fmt.Errorf("cert path must be specified when TLS is enabled")
	}

	if keyPath == "" {
		return nil, fmt.Errorf("cert key path must be specified when TLS is enabled")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading cert key pair: %w", err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error loading system certs: %w", err)
	}

	for _, caPath := range caPaths {
		caCertPEM, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("error reading CA cert in %s: %w", caPath, err)
		}

		if !certPool.AppendCertsFromPEM(caCertPEM) {
			return nil, fmt.Errorf("invalid CA cert in %s", caPath)
		}
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

func LoadCACerts(paths []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	if len(paths) == 0 {
		return certPool, nil
	}

	for _, path := range paths {
		certPEM, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading cert in %s: %w", path, err)
		}

		if !certPool.AppendCertsFromPEM(certPEM) {
			return nil, fmt.Errorf("invalid cert in %s", path)
		}
	}

	return certPool, nil
}

func SerializeCertPEMBytes(certPEMs [][]byte) ([]byte, error) {
	if len(certPEMs) == 0 {
		return []byte{}, nil
	}

	var allPEM bytes.Buffer
	for _, pemData := range certPEMs {
		if _, err := allPEM.Write(pemData); err != nil {
			return nil, fmt.Errorf("failed to write certificate data: %w", err)
		}
	}

	return allPEM.Bytes(), nil
}

func ParseJWK(rawKey string) (jwk.Key, error) {
	var key jwk.Key

	key, err := jwk.ParseKey([]byte(rawKey))
	if err != nil {
		return key, err
	}

	return key, nil
}
