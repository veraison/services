// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrKeyNotFound = errors.New("key not found")

func sanitizeKV(key, val string) error {
	if err := sanitizeK(key); err != nil {
		return err
	}

	if err := sanitizeV(val); err != nil {
		return err
	}

	return nil
}

func sanitizeK(key string) error {
	if key == "" {
		return errors.New("the supplied key is empty")
	}

	return nil
}

func sanitizeV(val string) error {
	var tmp interface{}

	if err := json.Unmarshal([]byte(val), &tmp); err != nil {
		return fmt.Errorf("the supplied val contains invalid JSON: %w", err)
	}

	return nil
}

func getMultiple(store IKVStore, keys []string) ([]string, error) {
	var vals []string
	var notFound []string

	for _, key := range keys {

		keyVals, err := store.Get(key)
		if err != nil {
			if errors.Is(err, ErrKeyNotFound) {
				notFound = append(notFound, fmt.Sprintf("%q", key))
			} else {
				return nil, err
			}
		}

		vals = append(vals, keyVals...)
	}

	var err error = nil
	if len(notFound) != 0 {
		err = fmt.Errorf("%w: %s", ErrKeyNotFound, strings.Join(notFound, ", "))
	}

	return vals, err
}
