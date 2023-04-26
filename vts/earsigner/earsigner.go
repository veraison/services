// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package earsigner

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

type Cfg struct {
	Alg string `mapstructure:"alg"`
	Key string `mapstructure:"key" config:"zerodefault"`
	Att string `mapstructure:"attester" config:"zerodefault"`
}

func New(v *viper.Viper, fs afero.Fs) (IEarSigner, error) {
	var cfg Cfg

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	// JWT is the only supported signing format for now
	// (CWT is not yet implemented in veraison/ear)
	es := &JWT{}

	if err := es.Init(cfg, fs); err != nil {
		return nil, err
	}

	return es, nil
}
