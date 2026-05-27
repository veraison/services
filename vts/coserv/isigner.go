// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
)

type ISigner interface {
	Sign(coserv coserv.Coserv) ([]byte, error)
	GetCoservSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error)
	GetAuthority() (*comid.CryptoKey, error)
	Close() error
}
