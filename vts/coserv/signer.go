// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

type SignerConfig struct {
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

func NewSigner(v *viper.Viper, fs afero.Fs) (ISigner, error) {
	cfg := SignerConfig{}

	configLoader := config.NewLoader(&cfg)
	if err := configLoader.LoadFromViper(v); err != nil {
		return nil, err
	}

	cs := COSESigner{}

	if err := cs.Init(cfg, fs); err != nil {
		return nil, err
	}

	return &cs, nil
}
