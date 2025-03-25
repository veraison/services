// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coservsigner

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/afero"
	"github.com/veraison/corim/coserv"
)

type ICoservSigner interface {
	Init(cfg Cfg, fs afero.Fs) error
	Sign(coserv coserv.Coserv) ([]byte, error)
	GetCoservSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error)
	Close() error
}
