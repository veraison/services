// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"fmt"
	"net/url"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

type Cfg struct {
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

func New(v *viper.Viper, fs afero.Fs) (IEarSigner, error) {
	var cfg Cfg

	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, err
	}

	keyUrl, err := url.Parse(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("invaid EAR signer key config: %w", err)
	}

	key, err := NewKeyLoader(fs).Load(keyUrl)
	if err != nil {
		return nil, fmt.Errorf("could not load EAR signer key: %w", err)
	}

	// JWT is the only supported signing format for now
	// (CWT is not yet implemented in veraison/ear)
	es := &JWT{}

	if err := es.Init(cfg, key); err != nil {
		return nil, err
	}

	return es, nil
}
