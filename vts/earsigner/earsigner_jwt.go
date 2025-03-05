// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/ear"
)

type JWT struct {
	Key interface{}
	Alg jwa.KeyAlgorithm
}

func (o *JWT) Init(cfg Cfg, key []byte) error {
	if err := o.setAlg(cfg.Alg); err != nil {
		return err
	}

	if err := o.loadKey(key); err != nil {
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

func (o JWT) GetEARSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error) {
	v, ok := o.Key.(jwk.Key)

	if ok != true {
		err := fmt.Errorf("error: failed conversion")
		return nil, nil, err
	}

	key, err := v.PublicKey()

	return o.Alg, key, err
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

func (o *JWT) loadKey(raw []byte) error {
	var (
		err error
		k   jwk.Key
	)

	if k, err = jwk.ParseKey(raw); err != nil {
		return fmt.Errorf("parsing signing key: %w", err)
	}

	o.Key = k

	return nil
}
