// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/config"
)

// Base64URLToBytes descdoes []byte for the provided string. If the string does
// not contains a valid base64url value (optionally, with padding), and error
// is returned.
func Base64URLToBytes(value string) ([]byte, error) {
	result, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return result, nil
	}

	return base64.URLEncoding.DecodeString(value)
}

// DecodePublicKeyPEM decodes a PEM encoded SubjectPublicKeyInfo.
func DecodePublicKeyPEM(key []byte) (crypto.PublicKey, error) {
	block, rest := pem.Decode(key)
	if block == nil {
		return nil, errors.New("could not extract trust anchor PEM block")
	}

	if len(rest) != 0 {
		return nil, errors.New("trailing data found after PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("unsupported key type: %q", block.Type)
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse public key: %w", err)
	}
	return pk, nil
}

// ToMapViaJSON converts any value into a map[string]any by transcoding it via
// JSON.
func ToMapViaJSON(val any) (map[string]any, error) {
	encoded, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	var ret map[string]any
	if err := json.Unmarshal(encoded, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// PluginNameFromScheme generates a plugin name from an attestations scheme
// name.
func PluginNameFromScheme(schemeName string) string {
	name := strings.ToLower(strings.ReplaceAll(schemeName, " ", "-"))
	return fmt.Sprintf("%s-scheme-plugin", name)
}

// CreateAttestationResult creates a new attestation result containing a single
// submod with the specified name.
func CreateAttestationResult(submodName string) *ear.AttestationResult {
	return ear.NewAttestationResult(submodName, config.Version, config.Developer)
}
