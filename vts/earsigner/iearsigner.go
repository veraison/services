// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/ear"
)

type IEarSigner interface {
	Init(cfg Cfg, key []byte) error
	Sign(earClaims ear.AttestationResult) ([]byte, error)
	GetEARSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error)
	Close() error
}
