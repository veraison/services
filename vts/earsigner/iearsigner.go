// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"crypto/sha256"
	"fmt"

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
	Att *Attestation
}

type Attestation struct {
	TEE      string
	UID      string
	Evidence []byte
}

func NewAttestation(tee string, evidence []byte) *Attestation {
	// automatically generate a cache id for the key+platform attestation
	uid := sha256.New()
	uid.Write(evidence)
	uid.Sum(nil)

	return &Attestation{
		TEE:      tee,
		UID:      fmt.Sprintf("%x", uid),
		Evidence: evidence,
	}
}
