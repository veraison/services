// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsementstore

import (
	"fmt"
	"maps"
	"time"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/config"
	"github.com/veraison/services/plugin"
	vtscoserv "github.com/veraison/services/vts/coserv"
)

// AggregateStoreParams aggregates different configurations to create
// a consolidated parameters map, that can be passed to the store plugin
// manager. The configurations come from store plugin configuration
// section, vts configuration section and coserv configuration section.
func AggregateStoreParams(
	pluginConfig map[string]*plugin.Parameters,
	vtsConfig *viper.Viper,
	coservContext *vtscoserv.Context,
) (map[string]*plugin.Parameters, error) {
	ret := maps.Clone(pluginConfig)

	commParams, err := CreateStoreCommonParams(coservContext)
	if err != nil {
		return nil, fmt.Errorf("could not extract common store params: %w", err)
	}

	activeStores, err := getActiveStores(vtsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get active-stores from vts config: %w", err)
	}

	// create parameter maps for plugins that do not have
	// config parameters in their plugin configuration
	for _, s := range activeStores {
		if _, ok := ret[s]; !ok {
			ret[s] = plugin.NewParameters()
		}
	}

	// broadcast common config to all plugins
	for _, v := range ret {
		if err := v.Merge(commParams); err != nil {
			return nil, fmt.Errorf("failed to aggregate store config: %w", err)
		}
	}
	return ret, nil
}

// Get list of active stores from the active-stores section of
// vts configuration.
func getActiveStores(vtsConf *viper.Viper) ([]string, error) {
	var list = struct {
		ActiveStores []string `mapstructure:"active-stores"`
	}{}
	loader := config.NewNonExclusiveLoader(&list)

	if err := loader.LoadFromViper(vtsConf); err != nil {
		return nil, err
	}
	return list.ActiveStores, nil
}

// private constants used for serializing and deserializing
// StoreCommonParams
var (
	fallbackAuthorityKey = "coserv-fallback-authority"
	maxExpiryKey         = "coserv-max-expiry"
)

// The configuration parameters common for all stores.
// These parameters are extracted from VTS configuration and
// broadcast to all store plugins.
type StoreCommonParams struct {
	FallbackAuthority *comid.CryptoKey
	MaxExpiry         time.Duration
}

// Create StoreCommonParams from a coservContext
func CreateStoreCommonParams(coservContext *vtscoserv.Context) (*plugin.Parameters, error) {
	if coservContext == nil {
		return nil, nil
	}
	if coservContext.FallbackAuthority == nil {
		return nil, fmt.Errorf("fallback authority is nil")
	}
	ret := &StoreCommonParams{
		coservContext.FallbackAuthority,
		coservContext.MaxExpiry,
	}
	return ret.ToParams()
}

// Deserialize plugin.Parameters into StoreCommonParams
func (o *StoreCommonParams) FromParams(params *plugin.Parameters) error {
	var key comid.CryptoKey

	auth, err := params.GetBytes(fallbackAuthorityKey)
	if err != nil {
		return err
	}

	exp, err := params.GetInt64(maxExpiryKey)
	if err != nil {
		return err
	}

	if err := (&key).UnmarshalCBOR(auth); err != nil {
		return err
	}

	o.FallbackAuthority = &key
	o.MaxExpiry = time.Duration(exp)
	return nil
}

// Serialize StoreCommonParams to plugin.Parameters
func (o *StoreCommonParams) ToParams() (*plugin.Parameters, error) {
	auth, err := o.FallbackAuthority.MarshalCBOR()
	if err != nil {
		return nil, err
	}
	exp := o.MaxExpiry.Nanoseconds()

	params := plugin.NewParameters().
		SetBytes(fallbackAuthorityKey, auth).
		SetInt64(maxExpiryKey, exp)

	return params, nil
}
