// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"go.uber.org/zap"
)

type cfg struct {
	Backend        string
	BackendConfigs map[string]interface{} `mapstructure:",remain"`
}

func (o cfg) Validate() error {
	supportedBackends := map[string]bool{
		"memory": true,
		"sql":    true,
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

func New(v *viper.Viper, logger *zap.SugaredLogger) (IKVStore, error) {
	var cfg cfg

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	var s IKVStore

	switch cfg.Backend {
	case "memory":
		s = &Memory{}
	case "sql":
		s = &SQL{}
	default:
		return nil, fmt.Errorf("backend %q is not supported", cfg.Backend)
	}

	if err := s.Init(v, logger); err != nil {
		return nil, err
	}

	return s, nil
}
