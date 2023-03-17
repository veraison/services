// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/afero"
	"github.com/veraison/ear"
)

type IEarSigner interface {
	Init(cfg Cfg, fs afero.Fs) error
	Sign(earClaims ear.AttestationResult) ([]byte, error)
	GetEARSigningPublicKeyEar() (jwk.Key, error)
	Close() error
}
