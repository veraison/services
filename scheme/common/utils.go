// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

func GetFieldsFromParts(parts *structpb.Struct) (map[string]*structpb.Value, error) {
	if parts == nil {
		return nil, errors.New("no parts found")
	}

	fields := parts.GetFields()
	if fields == nil {
		return nil, errors.New("no fields found")
	}

	return fields, nil
}

func GetMandatoryPathSegment(key string, fields map[string]*structpb.Value) (string, error) {
	v, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("mandatory %s is missing", key)
	}

	segment := v.GetStringValue()
	if segment == "" {
		return "", fmt.Errorf("mandatory %s is empty", key)
	}

	return segment, nil
}

func GetPubKeyFromPEMEncodedPubKey(ta []byte) (crypto.PublicKey, error) {
	block, rest := pem.Decode(ta)
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
