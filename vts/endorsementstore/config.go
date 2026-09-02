// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsementstore

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/plugin"
	vtscoserv "github.com/veraison/services/vts/coserv"
)

// Creates the plugin parameters map by combining the plugin specific
// parameters with the common CoSERV configuration
func CreateStoreParams(
	cfg *viper.Viper,
	coservCfg *vtscoserv.StoreConfig,
) (map[string]*plugin.Parameters, error) {
	var (
		err          error
		coservParams *plugin.Parameters
		pluginParams map[string]*plugin.Parameters
	)
	if cfg == nil {
		// caller should ensure that the config is non-nil
		panic("empty endorsement store configuration")
	}
	if cfg.Sub("coserv") != nil && coservCfg == nil {
		// plugins can assume that CoSERV API is disabled if CoSERV
		// configuration is not passed to them
		panic("invalid endorsement-store configuration for CoSERV")
	}

	activePlugins := cfg.GetStringSlice("active-plugins")
	if len(activePlugins) == 0 {
		return nil, fmt.Errorf("failed to read active store plugins list")
	}

	if p := cfg.Sub("plugin-parameters"); p != nil {
		pluginParams, err = plugin.ParametersMapFromViper(p, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read store plugin parameters: %w", err)
		}
	}

	if coservCfg == nil {
		// no CoSERV parameters to broadcast
		return pluginParams, nil
	}

	coservParams, err = coservCfg.ToParams()
	if err != nil {
		return nil, fmt.Errorf("invalid CoSERV parameters: %w", err)
	}

	if pluginParams == nil {
		pluginParams = make(map[string]*plugin.Parameters, len(activePlugins))
	}

	// broadcast coserv config to all plugins
	for _, s := range activePlugins {
		// create parameter maps for plugins that do not have
		// config parameters in their plugin configuration
		if _, ok := pluginParams[s]; !ok {
			pluginParams[s] = plugin.NewParameters()
		}
		if err := pluginParams[s].Merge(coservParams); err != nil {
			return nil, fmt.Errorf("failed to create store config: %w", err)
		}
	}

	return pluginParams, nil
}
