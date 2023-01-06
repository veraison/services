// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/afero"
	"github.com/veraison/ear"
)

type JWT struct {
	Key interface{}
	Alg jwa.KeyAlgorithm
}

func (o *JWT) Init(cfg Cfg, fs afero.Fs) error {
	if err := o.setAlg(cfg.Alg); err != nil {
		return err
	}

	if err := o.loadKey(fs, cfg.Key); err != nil {
		return err
	}

	// TODO(tho) optimisation: check that key and alg are compatible rather than
	// leaving it for when Sign is invoked

	return nil
}

func (o *JWT) Close() error {
	return nil
}

func (o JWT) Sign(earClaims ear.AttestationResult) ([]byte, error) {
	return earClaims.Sign(o.Alg, o.Key)
}

func (o *JWT) setAlg(alg string) error {
	for _, a := range jwa.SignatureAlgorithms() {
		if a.String() == alg {
			o.Alg = a
			return nil
		}
	}

	return fmt.Errorf(
		"%q is not a valid value for %q.  Supported algorithms: %s",
		alg, "ear-signer.alg", algList(),
	)
}

func algList() string {
	var l []string

	for _, a := range jwa.SignatureAlgorithms() {
		l = append(l, string(a))
	}

	return strings.Join(l, ", ")
}

func (o *JWT) loadKey(fs afero.Fs, keyFile string) error {
	var (
		err error
		b   []byte
		k   jwk.Key
	)

	if b, err = afero.ReadFile(fs, keyFile); err != nil {
		return fmt.Errorf("loading signing key from %q: %w", keyFile, err)
	}

	if k, err = jwk.ParseKey(b); err != nil {
		return fmt.Errorf("parsing signing key from %q: %w", keyFile, err)
	}

	o.Key = k

	return nil
}
