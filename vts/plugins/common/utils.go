// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
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
