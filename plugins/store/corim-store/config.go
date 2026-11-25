// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corim_store

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
)

type Config struct {
	DBMS     string                    `mapstructure:"dbms"`
	DSN      string                    `mapstructure:"dsn"`
	TraceSQL bool                      `mapstructure:"trace-sql" config:"zerodefault"`
	CoSERV   *endorsement.CoservConfig `config:"zerodefault"`
}

func ConfigFromParameters(params *plugin.Parameters, logger *zap.SugaredLogger) (*Config, error) {
	var (
		cfg       Config
		coservCfg endorsement.CoservConfig
	)

	logger.Debug("creating corim-store config from plugin parameters")
	loader := config.NewNonExclusiveLoader(&cfg)
	if err := loader.LoadFromMap(params.Map()); err != nil {
		return nil, err
	}

	// CoSERV configuration not being present is not an error
	if err := (&coservCfg).FromParams(params); err != nil {
		logger.Warnf("error converting parameters to coserv config: %v", err)
		cfg.CoSERV = nil
	} else {
		cfg.CoSERV = &coservCfg
	}

	return &cfg, nil
}

func (o *Config) StoreConfig() *store.Config {
	ret := store.NewConfig(
		o.DBMS,
		o.DSN,
		store.OptionRequireLabel,
		// Signatures of signed CoRIMs are verified as part of their
		// initial processing, prior to adding them to the store.
		store.OptionInsecure,
	)

	if o.TraceSQL {
		ret.WithOptions(store.OptionTraceSQL)
	}

	return ret
}

func (o *Config) CoservConfig() *endorsement.CoservConfig {
	return o.CoSERV
}

func New(v *viper.Viper, logger *zap.SugaredLogger) (*store.Store, error) {
	var cfg Config

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	logger.Debugf("connecting to %s store %s", cfg.DBMS, cfg.DSN)

	ret, err := store.Open(context.Background(), cfg.StoreConfig())
	if err != nil {
		return nil, err
	}

	// The store must be initialized before it may be used. In general, we
	// rely on the store being pointed to by DSN to be initialized prior
	// to starting the VTS. For in-memory store this can never be the case, so we
	// initialize it here.
	if strings.Contains(cfg.DSN, ":memory:") {
		if err := ret.Init(); err != nil {
			return nil, err
		}
	}

	return ret, nil
}
