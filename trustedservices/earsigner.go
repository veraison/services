// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/veraison/services/config"
)

type SignerCfg struct {
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

func NewJWTSigner(v *viper.Viper, fs afero.Fs) (IEarSigner, error) {
	var cfg SignerCfg

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	// JWT is the only supported signing format for now
	// (CWT is not yet implemented in veraison/ear)
	es := &JWTSigner{}

	if err := es.Init(cfg, fs); err != nil {
		return nil, err
	}

	return es, nil
}
