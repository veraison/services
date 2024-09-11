// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/veraison/ccatoken/platform"
	"github.com/veraison/ccatoken/realm"
	"github.com/veraison/psatoken"
)

type CcaPlatformWrapper struct {
	C platform.IClaims
}

func (o CcaPlatformWrapper) MarshalJSON() ([]byte, error) {
	return platform.ValidateAndEncodeClaimsToJSON(o.C)
}

type CcaRealmWrapper struct {
	C realm.IClaims
}

func (o CcaRealmWrapper) MarshalJSON() ([]byte, error) {
	return realm.ValidateAndEncodeClaimsToJSON(o.C)
}

type PsaPlatformWrapper struct {
	C psatoken.IClaims
}

func (o PsaPlatformWrapper) MarshalJSON() ([]byte, error) {
	return psatoken.ValidateAndEncodeClaimsToJSON(o.C)
}

type ClaimMapper interface {
	MarshalJSON() ([]byte, error)
}

func ClaimsToMap(mapper ClaimMapper) (map[string]interface{}, error) {
	data, err := mapper.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)

	return out, err
}

func MapToPSAClaims(in map[string]interface{}) (psatoken.IClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return psatoken.DecodeAndValidateClaimsFromJSON(data)
}

func MapToCCAPlatformClaims(in map[string]interface{}) (platform.IClaims, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return platform.DecodeAndValidateClaimsFromJSON(data)
}

func GetImplID(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}

	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Implementation ID for scheme: %s %w", scheme, err)
	}
	key := "impl-id"
	implID, ok := at[key].(string)
	if !ok {
		return "", fmt.Errorf("unable to get Implementation ID for scheme: %s", scheme)
	}
	return implID, nil
}

func GetInstID(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Instance ID: %w", err)
	}
	key := "inst-id"
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
