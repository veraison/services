// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/plugin"
)

func TestConfigStoreConfigFrom(t *testing.T) {
	params1 := plugin.NewParameters()
	params1.SetBytes(fallbackAuthorityKey, cryptoKeyCBOR())
	params1.SetInt64(maxExpiryKey, expiryInt())

	params2 := plugin.NewParameters()

	params3 := plugin.NewParameters()
	params3.SetBytes(fallbackAuthorityKey, cryptoKeyCBOR())

	params4 := plugin.NewParameters()
	params4.SetBytes(fallbackAuthorityKey, []byte{0xa2}) // invalid CBOR
	params4.SetInt64(maxExpiryKey, expiryInt())

	storePar1 := new(StoreConfig)
	err := storePar1.FromParams(params1)
	assert.NoError(t, err)
	assert.Equal(t, storePar1.MaxExpiry, expiry())
	assert.Equal(t, *cryptoKey(), *storePar1.Authority)

	storePar2 := new(StoreConfig)
	err = storePar2.FromParams(params2)
	assert.Error(t, err)

	storePar3 := new(StoreConfig)
	err = storePar3.FromParams(params3)
	assert.Error(t, err)

	storePar4 := new(StoreConfig)
	err = storePar4.FromParams(params4)
	assert.Error(t, err)
}

func TestConfigStoreConfigTo(t *testing.T) {
	cfg := StoreConfig{
		cryptoKey(),
		expiry(),
	}

	params, err := (&cfg).ToParams()

	assert.NoError(t, err)

	exp, err := params.GetInt64(maxExpiryKey)
	assert.NoError(t, err)
	assert.Equal(t, exp, expiryInt())

	auth, err := params.GetBytes(fallbackAuthorityKey)
	assert.NoError(t, err)
	assert.Equal(t, auth, cryptoKeyCBOR())
}

func cryptoKey() *comid.CryptoKey {
	k, err := comid.NewCryptoKeyTaggedBytes([]byte{0, 0, 0, 0})
	if err != nil {
		panic(err)
	}
	return k
}

func cryptoKeyCBOR() []byte {
	bytes, err := cryptoKey().MarshalCBOR()
	if err != nil {
		panic(err)
	}
	return bytes
}

func expiryInt() int64 {
	return expiry().Nanoseconds()
}

func expiry() time.Duration {
	d, err := time.ParseDuration("300s")
	if err != nil {
		panic(err)
	}
	return d
}
