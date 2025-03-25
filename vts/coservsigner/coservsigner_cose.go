// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coservsigner

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/afero"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/go-cose"
	"github.com/veraison/services/vts/earsigner"
)

type COSESigner struct {
	// JWK is used for interchange of key material (reading from configuration
	// and exporting to the discovery API)
	Key    jwk.Key
	rawkey []byte
	Alg    jwa.KeyAlgorithm

	// COSE is used for the actual signing of CoSERVs
	Signer cose.Signer
}

func (o *COSESigner) Init(cfg Cfg, fs afero.Fs) error {
	keyUrl, err := url.Parse(cfg.Key)
	if err != nil {
		return fmt.Errorf("parsing CoSERV signer key from configuration: %w", err)
	}

	key, err := earsigner.NewKeyLoader(fs).Load(keyUrl)
	if err != nil {
		return fmt.Errorf("loading CoSERV signer key: %w", err)
	}

	if err := o.parseKey(key); err != nil {
		return fmt.Errorf("parsing CoSERV signer key: %w", err)
	}

	if err := o.setAlg(cfg.Alg); err != nil {
		return fmt.Errorf("setting CoSERV signer algorithm: %w", err)
	}

	// instantiate the cose.Signer from the loaded key and alg
	if err := o.initSigner(); err != nil {
		return fmt.Errorf("creating COSE signer: %w", err)
	}

	return nil
}

func (o *COSESigner) Close() error {
	return nil
}

func (o COSESigner) Sign(c coserv.Coserv) ([]byte, error) {
	return c.Sign(o.Signer)
}

func (o COSESigner) GetCoservSigningPublicKey() (jwa.KeyAlgorithm, jwk.Key, error) {
	key, err := o.Key.PublicKey()
	if err != nil {
		return nil, nil, fmt.Errorf("getting public key from CoSERV signer key: %w", err)
	}

	return o.Alg, key, nil
}

func (o *COSESigner) setAlg(alg string) error {
	for _, a := range jwa.SignatureAlgorithms() {
		if a.String() == alg {
			o.Alg = a
			return nil
		}
	}

	return fmt.Errorf(
		"%q is not a valid value for %q.  Supported algorithms: %s",
		alg, "coserv-signer.alg", algList(),
	)
}

func algList() string {
	var l []string

	for _, a := range jwa.SignatureAlgorithms() {
		l = append(l, string(a))
	}

	return strings.Join(l, ", ")
}

func (o *COSESigner) parseKey(raw []byte) error {
	k, err := jwk.ParseKey(raw)
	if err != nil {
		return err
	}

	if _, ok := k.(jwk.AsymmetricKey); !ok {
		return fmt.Errorf("expected asymmetric key, got %T", k)
	}

	if !k.(jwk.AsymmetricKey).IsPrivate() {
		return errors.New("expected private key, got public key")
	}

	o.Key = k
	o.rawkey = raw

	return nil
}

func jwkToMap(k []byte) (map[string]string, error) {
	var x map[string]interface{}

	if err := json.Unmarshal(k, &x); err != nil {
		return nil, fmt.Errorf("getting key map from CoSERV signer key: %w", err)
	}

	y := make(map[string]string, len(x))

	for k, v := range x {
		if s, ok := v.(string); ok {
			y[k] = s
		} else {
			return nil, fmt.Errorf("cannot convert value for k %q: expecting string, got %T", k, v)
		}
	}

	return y, nil
}

func (o *COSESigner) initSigner() error {
	k, err := jwkToMap(o.rawkey)
	if err != nil {
		return fmt.Errorf("internalising JWK: %w", err)
	}

	cryptoSigner, err := getCryptoSigner(k)
	if err != nil {
		return fmt.Errorf("mapping JWK to crypto.Signer: %w", err)
	}

	alg, err := getAlg(o.Alg)
	if err != nil {
		return fmt.Errorf("mapping JWK to crypto.Signer: %w", err)
	}

	coseSigner, err := cose.NewSigner(alg, cryptoSigner)
	if err != nil {
		return fmt.Errorf("mapping JWK to crypto.Signer: %w", err)
	}

	o.Signer = coseSigner

	return nil
}

func getAlg(alg jwa.KeyAlgorithm) (cose.Algorithm, error) {
	switch alg {
	case jwa.ES256:
		return cose.AlgorithmES256, nil
	case jwa.ES384:
		return cose.AlgorithmES384, nil
	case jwa.ES512:
		return cose.AlgorithmES512, nil
	}
	return 0, fmt.Errorf("unsupported signing algorithm: %s", alg)
}

func getCryptoSigner(k map[string]string) (crypto.Signer, error) {
	if k["kty"] == "EC" {
		var c elliptic.Curve
		switch k["crv"] {
		case "P-256":
			c = elliptic.P256()
		case "P-384":
			c = elliptic.P384()
		case "P-521":
			c = elliptic.P521()
		default:
			return nil, errors.New("unsupported EC curve: " + k["crv"])
		}

		x, err := base64ToBigInt(k["x"])
		if err != nil {
			return nil, fmt.Errorf("decoding x coordinate: %w", err)
		}
		y, err := base64ToBigInt(k["y"])
		if err != nil {
			return nil, fmt.Errorf("decoding y coordinate: %w", err)
		}
		d, err := base64ToBigInt(k["d"])
		if err != nil {
			return nil, fmt.Errorf("decoding d coordinate: %w", err)
		}

		pkey := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				X:     x,
				Y:     y,
				Curve: c,
			},
			D: d,
		}
		return pkey, nil
	}

	return nil, errors.New("unsupported key type: " + k["kty"])
}

func base64ToBigInt(s string) (*big.Int, error) {
	val, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(val), nil
}
