// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/veraison/psatoken"
)

type ClaimMapper interface {
	ToJSON() ([]byte, error)
}

func ClaimsToMap(mapper ClaimMapper) (map[string]interface{}, error) {
	data, err := mapper.ToJSON()
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	return out, err
}

func MapToClaims(in map[string]interface{}) (psatoken.IClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return psatoken.DecodeJSONClaims(data)
}

func GetImplID(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Implementation ID: %w", err)
	}
	key := scheme + ".impl-id"
	implID, ok := at[key].(string)
	if !ok {
		return "", errors.New("unable to get Implementation ID")
	}
	return implID, nil
}

func GetInstID(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Instance ID: %w", err)
	}
	key := scheme + ".inst-id"
	instID, ok := at[key].(string)
	if !ok {
		return "", errors.New("unable to get Instance ID")
	}
	return instID, nil
}

// DecodePemSubjectPubKeyInfo decodes a PEM encoded SubjectPublicKeyInfo
func DecodePemSubjectPubKeyInfo(key []byte) (crypto.PublicKey, error) {
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
