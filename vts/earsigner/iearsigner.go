// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"github.com/spf13/afero"
	"github.com/veraison/ear"
)

type IEarSigner interface {
	Init(cfg Cfg, fs afero.Fs) error
	Sign(earClaims ear.AttestationResult) ([]byte, error)
	Close() error
}
