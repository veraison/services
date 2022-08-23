// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import "errors"

var ErrStoreNotFound = errors.New("store not found")
var ErrStoreAlreadyExists = errors.New("store already exists")

// IReader specifies an interface for a config file reader. A config reader
// parses configuration contents and uses it to populate one or more Stores.
type IReader interface {
	// Read reads configuration from the specified buffer, returning the
	// total number of bytes read, and any errors encountered.
	Read (buf []byte) (int, error)

	// ReadFile reads configuration from a file located at the specified
	// path, returning the total number of bytes read, and any errors
	// encountered.
	ReadFile(path string) (int, error)

	// GetStores returns a map of Stores parsed using Read or ReadFile,
	// keyed by the names under which they appeared in the configuration.
	GetStores() map[string]Store

	// GetStore returns a Store that was loaded under the specified name.
	// If no such Store exists, an error is returned.
	GetStore(name string) (Store, error)

	// MustGetStore returns a Store that was loaded under the specified
	// name. If no such store exists, an empty store is returned instead.
	MustGetStore(name string) Store
}
