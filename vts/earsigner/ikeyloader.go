// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import "net/url"

// IKeyLoader defines the interface for loading signing keys
type IKeyLoader interface {
	// Load the signing key from the specified location, returning it as a
	// []byte, if successful, or an error if not.
	Load(location *url.URL) ([]byte, error)
}
