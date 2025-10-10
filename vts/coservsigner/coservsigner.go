// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coservsigner

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

type Cfg struct {
	Use bool   `mapstructure:"use"`
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

func New(v *viper.Viper, fs afero.Fs) (ICoservSigner, error) {
	if !v.GetBool("use") {
		return nil, nil
	}

	cfg := Cfg{}

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
