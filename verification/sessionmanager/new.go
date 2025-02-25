// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)


const DefaultBackend = "ttlcache"

type cfg struct {
	Backend        string
	BackendConfigs map[string]interface{} `mapstructure:",remain"`
}

func (o cfg) Validate() error {
	supportedBackends := map[string]bool{
		"ttlcache": true,
		"memcached": true,
	}

	var unexpected []string
	for k := range o.BackendConfigs {
		if _, ok := supportedBackends[k]; !ok {
			unexpected = append(unexpected, k)
		}
	}

	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		return fmt.Errorf("unexpected directives: %s", strings.Join(unexpected, ", "))
	}

	return nil
}

func New(v *viper.Viper) (ISessionManager, error) {
	cfg := cfg{
		Backend: DefaultBackend,
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	var sm ISessionManager
	switch cfg.Backend {
	case "ttlcache":
		sm = NewTTLCache()
	case "memcached":
		sm = NewMemcached()
	default:
		return nil, fmt.Errorf("backend %q is not supported", cfg.Backend)
	}

	if err := sm.Init(v.Sub(cfg.Backend)); err != nil {
		return nil, err
	}

	return sm, nil
}
