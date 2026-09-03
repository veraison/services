// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package corimstore

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/config"
	vtsstore "github.com/veraison/services/vts/endorsementstore"
	"go.uber.org/zap"
)

type Config struct {
	DBMS     string `mapstructure:"dbms"`
	DSN      string `mapstructure:"dsn"`
	TraceSQL bool   `mapstructure:"trace-sql" config:"zerodefault"`
	*vtsstore.StoreCommonParams `config:"zerodefault"`
}

func ConfigFromParameters(params *plugin.Parameters, logger *zap.SugaredLogger) (*Config, error) {
	var (
		cfg Config
		coservCfg vtsstore.StoreCommonParams
	)

	logger.Debug("creating corim-store config from plugin parameters")
	loader := config.NewNonExclusiveLoader(&cfg)
	if err := loader.LoadFromMap(params.Map()); err != nil {
		return nil, err
	}

	// CoSERV configuration not being present is not an error
	if err := (&coservCfg).FromParams(params); err != nil {
		logger.Warnf("error converting parameters to coserv config: %v", err)
		cfg.StoreCommonParams = nil
	} else {
		cfg.StoreCommonParams = &coservCfg
	}

	return &cfg, nil
}

func (o *Config) StoreConfig() *corimstore.Config {
	ret := corimstore.NewConfig(
		o.DBMS,
		o.DSN,
		corimstore.OptionRequireLabel,
		// Signatures of signed CoRIMs are verified as part of their
		// initial processing, prior to adding them to the store.
		corimstore.OptionInsecure,
	)

	if o.TraceSQL {
		ret.WithOptions(corimstore.OptionTraceSQL)
	}

	return ret
}

func (o *Config) CoservConfig() *vtsstore.StoreCommonParams {
	return o.StoreCommonParams
}

func New(v *viper.Viper, logger *zap.SugaredLogger) (*corimstore.Store, error) {
	var cfg Config

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	logger.Debugf("connecting to %s store %s", cfg.DBMS, cfg.DSN)

	store, err := corimstore.Open(context.Background(), cfg.StoreConfig())
	if err != nil {
		return nil, err
	}

	// The store must be initialized before it may be used. In general, we
	// rely on the store being pointed to by DSN to be initialized prior
	// to starting the VTS. For in-memory store this can never be the case, so we
	// initialize it here.
	if strings.Contains(cfg.DSN, ":memory:") {
		if err := store.Init(); err != nil {
			return nil, err
		}
	}

	return store, nil
}
