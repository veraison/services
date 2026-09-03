// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsementstore

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/vts/coserv"
)

func TestConfigAggregateStoreParams(t *testing.T) {
	pluginCfg := make(map[string]*plugin.Parameters, 2)
	pl1 := plugin.NewParameters().SetString("p1_cfg", "some-cfg")
	pl2 := plugin.NewParameters()
	pluginCfg["pl1"] = pl1
	pluginCfg["pl2"] = pl2

	vtsCfg := vtsCfgGood()

	coservCtx := &coserv.Context{
		Signer:            nil,
		FallbackAuthority: cryptoKey(),
		MaxExpiry:         expiry(),
	}

	aggrCfg, err := AggregateStoreParams(pluginCfg, vtsCfg, coservCtx)
	assert.NoError(t, err)
	stores := slices.Collect(maps.Keys(aggrCfg))
	slices.Sort(stores)
	assert.Equal(t, stores, storeList())

	assert.Equal(t, aggrCfg["pl1"].MustGetString("p1_cfg"), "some-cfg")

	for _, cfg := range pluginCfg {
		assert.Equal(t, cfg.MustGetBytes(fallbackAuthorityKey), cryptoKeyCBOR())
		assert.Equal(t, cfg.MustGetInt64(maxExpiryKey), expiryInt())
	}

	_, err = AggregateStoreParams(pluginCfg, vtsCfgBad(), coservCtx)
	assert.Error(t, err)
}

func TestCreateStoreCommonParams(t *testing.T) {
	ctx1 := &coserv.Context{
		Signer:            nil,
		FallbackAuthority: nil,
		MaxExpiry:         expiry(),
	}

	ctx2 := &coserv.Context{
		Signer:            nil,
		FallbackAuthority: cryptoKey(),
		MaxExpiry:         expiry(),
	}

	var ctx3 *coserv.Context

	_, e1 := CreateStoreCommonParams(ctx1)
	params, e2 := CreateStoreCommonParams(ctx2)
	empty, e3 := CreateStoreCommonParams(ctx3)

	assert.Error(t, e1)
	assert.NoError(t, e2)
	assert.NoError(t, e3)
	assert.Nil(t, empty)

	exp, err := params.GetInt64(maxExpiryKey)
	assert.NoError(t, err)
	auth, err := params.GetBytes(fallbackAuthorityKey)
	assert.NoError(t, err)
	assert.Equal(t, exp, expiryInt())
	assert.Equal(t, auth, cryptoKeyCBOR())
}

func TestConfigGetActiveStores(t *testing.T) {
	stores := []string{"foo", "bar"}
	cfg := viper.New()
	cfg.Set("active-stores", stores)

	list, err := getActiveStores(cfg)
	assert.NoError(t, err)
	assert.Equal(t, list, stores)
}

func TestConfigStoreCommonParamsFrom(t *testing.T) {
	params1 := plugin.NewParameters()
	params1.SetBytes(fallbackAuthorityKey, cryptoKeyCBOR())
	params1.SetInt64(maxExpiryKey, expiryInt())

	params2 := plugin.NewParameters()

	params3 := plugin.NewParameters()
	params3.SetBytes(fallbackAuthorityKey, cryptoKeyCBOR())

	params4 := plugin.NewParameters()
	params4.SetBytes(fallbackAuthorityKey, []byte{0xa2})
	params4.SetInt64(maxExpiryKey, expiryInt())

	storePar1 := new(StoreCommonParams)
	err := storePar1.FromParams(params1)
	assert.NoError(t, err)
	assert.Equal(t, storePar1.MaxExpiry, expiry())
	assert.Equal(t, *cryptoKey(), *storePar1.FallbackAuthority)

	storePar2 := new(StoreCommonParams)
	err = storePar2.FromParams(params2)
	assert.Error(t, err)

	storePar3 := new(StoreCommonParams)
	err = storePar3.FromParams(params3)
	assert.Error(t, err)

	storePar4 := new(StoreCommonParams)
	err = storePar4.FromParams(params4)
	assert.Error(t, err)
}

func TestConfigStoreCommonParamsTo(t *testing.T) {
	cfg := StoreCommonParams{
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

func vtsCfgGood() *viper.Viper {
	cfg := viper.New()
	cfg.Set("active-stores", storeList())
	return cfg
}

func vtsCfgBad() *viper.Viper {
	cfg := viper.New()
	cfg.Set("some-other", 100)
	return cfg
}

func storeList() []string {
	return []string{"pl1", "pl2", "pl3"}
}
