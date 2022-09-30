// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"os"

	"github.com/setrofim/viper"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:     "polcli",
		Short:   "policy management client",
		Version: "0.0.1",
	}

	dsnFromFlag string
)

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&dsnFromFlag, "store", "s", "",
		"policy store datasource (only used for sql backend).")

	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(delCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		wd, err := os.Getwd()
		cobra.CheckErr(err)

		viper.AddConfigPath(wd)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetDefault("po-store.backend", "sql")
	viper.SetDefault("po-store.sql.driver", "sqlite3")
	viper.SetDefault("po-store.sql.datasource", "po-store.sql")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// If there is no config file, use the defaults set above.
		err = nil
	}

	cobra.CheckErr(err)

	if dsnFromFlag != "" {
		viper.Set("po-store.sql.datasource", dsnFromFlag)
	}
}
