// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// IKVStore is the interface to a key-value store. Keys and values are both
// strings. A key can be associated with multiple values.
type IKVStore interface {
	// Init initializes the store. The parameters expected inside
	// viper.Viper are implementation-specific -- please see the
	// documentation for the implementation you're using.
	Init(v *viper.Viper, logger *zap.SugaredLogger) error

	// Close the store, shutting down the underlying connection (if one
	// exists in the implementation), and disallowing any further
	// operations.
	Close() error

	// Setup a new store for use. What this actually entails is  specific
	// to a backend.
	Setup() error

	// Get returns a []string of values for the specified key. If the
	// specified key is not in the store, a ErrKeyNotFound is returned. The
	// values are in the order they were added, with the most recent value
	// last.
	Get(key string) ([]string, error)

	// GetKeys returns a []string of keys currently set in the store.
	GetKeys() ([]string, error)

	// Set the specified key to the specified value, discarding any
	// existing values.
	Set(key, val string) error

	// Del removes the specified key from the store, discarding its
	// associated values. If the key does not exist, ErrKeyNotFound will be
	// returned.
	Del(key string) error

	// Add the specified value to the specified key. If the key does
	// not already exist, this behaves like Set. If the key exists, the
	// specified val is appended to the existing value(s).
	Add(key, val string) error
}
