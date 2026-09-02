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

func TestCreateStoreParams(t *testing.T) {
	cfg1 := viper.New()
	cfg1.Set("coserv", coservSub())
	cfg1.Set("active-plugins", activePluginsSub())
	cfg1.Set("plugin-parameters", pluginParamsSub())

	cfg2 := viper.New()
	cfg2.Set("active-plugins", activePluginsSub())
	cfg2.Set("plugin-parameters", pluginParamsSub())

	cfg3 := viper.New()
	cfg3.Set("active-plugins", activePluginsSub())

	cfg4 := viper.New()
	cfg4.Set("coserv", coservSub())
	cfg4.Set("plugin-parameters", pluginParamsSub())

	cfg5 := viper.New()
	cfg5.Set("active-plugins", coservSub())

	cfg6 := viper.New()
	cfg6.Set("active-plugins", activePluginsSub())
	cfg6.Set("plugin-parameters", coservSub())

	cfg7 := viper.New()
	cfg7.Set("active-plugins", activePluginsSub())
	cfg7.Set("coserv", coservSub())

	coservCfg := &coserv.StoreConfig{
		Authority: cryptoKey(),
		MaxExpiry: expiry(),
	}

	assert.Panics(t, func() { CreateStoreParams(cfg1, nil) })      //nolint:errcheck
	assert.Panics(t, func() { CreateStoreParams(nil, coservCfg) }) //nolint:errcheck

	var (
		pm  map[string]*plugin.Parameters
		err error
	)
	pm, err = CreateStoreParams(cfg1, coservCfg)

	assert.NoError(t, err)
	assert.ElementsMatch(t, slices.Collect(maps.Keys(pm)), storeList())
	assert.Equal(t, pm["pl1"].MustGetString("pl1-cf1"), "val1")

	for _, cfg := range pm {
		var storeCfg coserv.StoreConfig
		assert.NoError(t, (&storeCfg).FromParams(cfg))
		assert.Equal(t, storeCfg.Authority, cryptoKey())
		assert.Equal(t, storeCfg.MaxExpiry, expiry())
	}

	_, err = CreateStoreParams(cfg2, coservCfg)
	assert.NoError(t, err)

	_, err = CreateStoreParams(cfg3, coservCfg)
	assert.NoError(t, err)

	_, err = CreateStoreParams(cfg5, coservCfg)
	assert.Error(t, err)

	_, err = CreateStoreParams(cfg6, coservCfg)
	assert.Error(t, err)

	_, err = CreateStoreParams(cfg7, coservCfg)
	assert.NoError(t, err)
}

func coservSub() map[string]any {
	m := map[string]any{
		"signer": map[string]any{
			"alg": "xx",
			"key": "yy",
		},
		"max-expiry": "1m",
	}
	return m
}

func activePluginsSub() []any {
	return []any{"pl1", "pl2"}
}

func pluginParamsSub() map[string]any {
	return map[string]any{
		"pl1": map[string]any{
			"pl1-cf1": "val1",
			"pl1-cf2": "val2",
		},
	}
}

func cryptoKey() *comid.CryptoKey {
	k, err := comid.NewCryptoKeyTaggedBytes([]byte{0, 0, 0, 0})
	if err != nil {
		panic(err)
	}
	return k
}

func expiry() time.Duration {
	d, err := time.ParseDuration("300s")
	if err != nil {
		panic(err)
	}
	return d
}

func storeList() []string {
	return []string{"pl1", "pl2"}
}
