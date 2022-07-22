// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"errors"
	"fmt"
)

type Store map[string]interface{}

var (
	ErrMissingDirective = errors.New("directive not found")
	ErrInvalidDirective = errors.New("invalidly specified directive")
)

// GetString returns the value of a directive from a configuration store. If
// supplied, a default value is returned in case the directive is not found. An
// error is returned if the value is not of type string or if the directive is
// not found and no default has been specified.
func GetString(store Store, directive string, dflt *string) (string, error) {
	v, ok := store[directive]
	if !ok {
		if dflt == nil {
			return "", fmt.Errorf("%q %w", directive, ErrMissingDirective)
		}
		return *dflt, nil
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w %q: want string, got %T", ErrInvalidDirective, directive, v)
	}

	return s, nil
}
