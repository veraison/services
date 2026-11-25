// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/util"
)

type EarSignerConfig struct {
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

type IEarSigner interface {
	Init(cfg EarSignerConfig, key []byte) error
	Sign(earClaims ear.AttestationResult) ([]byte, error)
	GetEARSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error)
	Close() error
}

func NewEarSigner(v *viper.Viper, fs afero.Fs) (IEarSigner, error) {
	var cfg EarSignerConfig

	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, err
	}

	keyUrl, err := url.Parse(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("invaid EAR signer key config: %w", err)
	}

	key, err := util.NewKeyLoader(fs).Load(keyUrl)
	if err != nil {
		return nil, fmt.Errorf("could not load EAR signer key: %w", err)
	}

	// JWT is the only supported signing format for now
	// (CWT is not yet implemented in veraison/ear)
	es := &EarSignerJWT{}

	if err := es.Init(cfg, key); err != nil {
		return nil, err
	}

	return es, nil
}

type EarSignerJWT struct {
	Key interface{}
	Alg jwa.KeyAlgorithm
}

func (o *EarSignerJWT) Init(cfg EarSignerConfig, key []byte) error {
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

func (o *EarSignerJWT) Close() error {
	return nil
}

func (o EarSignerJWT) Sign(earClaims ear.AttestationResult) ([]byte, error) {
	return earClaims.Sign(o.Alg, o.Key)
}

func (o EarSignerJWT) GetEARSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error) {
	v, ok := o.Key.(jwk.Key)

	if !ok {
		err := fmt.Errorf("error: failed conversion")
		return nil, nil, err
	}

	key, err := v.PublicKey()

	return o.Alg, key, err
}

func (o *EarSignerJWT) setAlg(alg string) error {
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
	jwaAlgs := jwa.SignatureAlgorithms()
	l := make([]string, len(jwaAlgs))

	for i, a := range jwaAlgs {
		l[i] = a.String()
	}

	return strings.Join(l, ", ")
}

func (o *EarSignerJWT) loadKey(raw []byte) error {
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
