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
	ErrUnsupported = errors.New("store does not support the operation")
	ErrNotFound    = errors.New("not found")
)

type IEndorsementStorePlugin interface {
	plugin.IPluggable

	IEndorsementStore
}

type IEndorsementStore interface {
	IEndorsementStoreReader

	IEndorsementStoreWriter

	Close() error
}

type IEndorsementStoreReader interface {
	GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error)

	GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error)

	// TODO: how to pass authority:
	// perhaps authority should be passed while initializing store
	ExecuteCoservQuery(mediaType, query string) (*coserv.Coserv, error)
}

type IEndorsementStoreWriter interface {
	AddCorimBytes(data []byte, scheme string, activate bool) error
}
