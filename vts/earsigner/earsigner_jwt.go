// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
	Tee *TEE
}

func (o *JWT) Init(cfg Cfg, fs afero.Fs) error {
	if err := o.setAlg(cfg.Alg); err != nil {
		return err
	}

	if cfg.Key == "" {
		// generate a new key pair if no ear-signer.key was supplied
		switch o.Alg {
		// TODO(tho) add other curves/algorithms
		case jwa.ES256:
			key, err := generateECDSAKey(jwa.P256)
			if err != nil {
				return fmt.Errorf("generating %v key: %w", o.Alg, err)
			}
			o.Key = key
		default:
			return fmt.Errorf("unsupported algorithm: %v", o.Alg)
		}

	} else if err := o.loadKey(fs, cfg.Key); err != nil {
		return err
	}

	// if requested, try and get an attestation for the signing key
	if att := cfg.Att; att != "" {
		k, err := getPK(o.Key)
		if err != nil {
			return err
		}

		switch att {
		case "nitro":
			b, err := nitroAttest(k)
			if err != nil {
				return fmt.Errorf("attesting EAR signing key failed: %w", err)
			}
			o.Tee = &TEE{Name: att, Evidence: b}
		default:
			return fmt.Errorf("unsupported attester type: %q", att)
		}
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

func getPK(k interface{}) (jwk.Key, error) {
	v, ok := k.(jwk.Key)
	if !ok {
		return nil, fmt.Errorf("failed conversion to JWK key")
	}

	return v.PublicKey()
}

func (o JWT) GetPublicKeyInfo() (PublicKeyInfo, error) {
	var (
		key jwk.Key
		err error
	)

	if key, err = getPK(o.Key); err != nil {
		return PublicKeyInfo{}, err
	}

	return PublicKeyInfo{
		Alg: o.Alg,
		Key: key,
		Tee: o.Tee,
	}, nil
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

func generateECDSAKey(alg jwa.EllipticCurveAlgorithm) (jwk.Key, error) {
	var crv elliptic.Curve

	if tmp, ok := CurveForAlgorithm(alg); ok {
		crv = tmp
	} else {
		return nil, fmt.Errorf("invalid curve algorithm %s", alg)
	}

	key, err := ecdsa.GenerateKey(crv, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating %v key: %w", alg, err)
	}

	return jwk.FromRaw(key)
}

func init() {
	RegisterCurve(elliptic.P256(), jwa.P256)
}
