// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

// ICoservProxyHandler defines the interface for CoSERV translation plugins
type ICoservProxyHandler interface {
	plugin.IPluggable

	// GetEndorsements adds the "result set" to the input "query" CoSERV.
	// The input query CoSERV is base64url-encoded.
	// In case of a failure an error is returned and the CoSERV is nil.
	GetEndorsements(tenantID string, query string) ([]byte, error)
}
