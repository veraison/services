// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

func SliceToListValue[T any](input []T) (*structpb.ListValue, error) {
	vals := make([]any, len(input))
	for i, v := range input {
		vals[i] = v
	}

	return structpb.NewList(vals)
}

func ListValueToSlice[T any](input *structpb.ListValue) ([]T, error) {
	vals := input.AsSlice()
	ret := make([]T, len(vals))
	for i, v := range vals {
		out, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", *new(T), v)
		}

		ret[i] = out
	}

	return ret, nil
}
