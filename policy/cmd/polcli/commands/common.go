// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"fmt"

	"github.com/setrofim/viper"
	"github.com/spf13/cobra"
	"github.com/veraison/services/config"
	"github.com/veraison/services/policy"
)

var (
	rawConfig        *viper.Viper
	store            *policy.Store
	storeDsnFromFlag string

	storeDefaults = map[string]string{
		"backend":        "sql",
		"sql.driver":     "sqlite3",
		"sql.datasource": "po-store.sql",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	rawConfig, err = config.ReadRawConfig(cfgFile, true)
	cobra.CheckErr(err)
}

func initPolicyStore(cmd *cobra.Command, args []string) error {
	cfg := rawConfig.Sub("po-store")
	for k, v := range storeDefaults {
		cfg.SetDefault(k, v)
	}

	// if store location has been specified with --store flag, set it as
	// the datasource for the selected backend.
	if storeDsnFromFlag != "" {
		cfg.Set(fmt.Sprintf("%s.datasource", cfg.GetString("backend")), storeDsnFromFlag)
	}

	var err error

	store, err = policy.NewStore(cfg)
	if err != nil {
		return fmt.Errorf("could not initialize policy store: %w", err)
	}

	return nil
}

func finiPolicyStore(cmd *cobra.Command, args []string) error {
	if store != nil {
		store.Close()
	}

	return nil
}
