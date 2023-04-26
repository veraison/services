// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/afero"
	"github.com/veraison/ear"
)

type IEarSigner interface {
	Init(cfg Cfg, fs afero.Fs) error
	Sign(earClaims ear.AttestationResult) ([]byte, error)
	GetPublicKeyInfo() (PublicKeyInfo, error)
	Close() error
}

type PublicKeyInfo struct {
	Alg jwa.KeyAlgorithm
	Key jwk.Key
	Tee *TEE
}

type TEE struct {
	Name     string
	Evidence []byte
}
