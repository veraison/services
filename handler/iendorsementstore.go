// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/plugin"
)

var (
	// When the store does not support the operation.
	// Example 1: store supportes scheme A and C, but
	// endorsements for scheme B is requested.
	//
	// Example 2: store does not support CoSERV but the
	// ExecuteCoservQuery is called.
	ErrUnsupported = errors.New("store does not support the operation")

	// Not found in store. This means the store supports the operation
	// and the attestation scheme, but no matching values were found
	// in the store.
	ErrNotFound = errors.New("not found in store")
)

type IEndorsementStorePlugin interface {
	plugin.IPluggable

	IEndorsementStore
}

type IEndorsementStore interface {
	GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error)

	GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error)

	ExecuteCoservQuery(profile, query string) (*coserv.Coserv, error)

	AddCorimBytes(data []byte, scheme string, activate bool) error
}
