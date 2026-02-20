// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package store

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/services/config"
	"go.uber.org/zap"
)

type Config struct {
	DBMS     string `mapstructure:"dbms"`
	DSN      string `mapstructure:"dsn"`
	TraceSQL bool   `mapstructure:"trace-sql" config:"zerodefault"`
}

func (o *Config) StoreConfig() *corimstore.Config {
	ret := corimstore.NewConfig(o.DBMS, o.DSN, corimstore.OptionRequireLabel)

	if o.TraceSQL {
		ret.WithOptions(corimstore.OptionTraceSQL)
	}

	return ret
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

	// The store must be innitialized before it may be used. In general, we
	// rely on the store being pointed to by DSN to be innitialized prior
	// to starting the VTS. For in-memory store this can never be the case, so we
	// initialize it here.
	if strings.Contains(cfg.DSN, ":memory:") {
		if err := store.Init(); err != nil {
			return nil, err
		}
	}

	return store, nil
}
