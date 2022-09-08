// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import "github.com/veraison/services/config"

// IKVStore is the interface to a key-value store. Keys and values are both
// strings. A key can be associated with multiple values.
type IKVStore interface {
	// Init initializes the store. The parameters of the config.Store are
	// implementation-specific -- please see the documentation for the
	// implementation you're using.
	Init(cfg config.Store) error

	// Close the store, shutting down the underlying connection (if one
	// exists in the implementation), and disallowing any further
	// operations.
	Close() error

	// Get returns a []string of values for the specified key. If the
	// specified key is not in the store, a ErrKeyNotFound is returned.
	Get(key string) ([]string, error)

	// Set the specified key to the specified value, discarding any
	// existing values.
	Set(key, val string) error

	// Del removes the specfied key from the store, discarding its
	// associated values.
	Del(key string) error

	// Add the specified value to the specified key. If the key does
	// not already exist, this behaves like Set. If the key exists, the
	// specified val is appended to the existing value(s).
	Add(key, val string) error
}
