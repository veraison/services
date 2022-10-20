// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:     "polcli",
		Short:   "policy management client",
		Version: "0.0.1",
	}
)

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&storeDsnFromFlag, "store", "s", "",
		"policy store datasource (only used for sql backend).")

	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(delCmd)
}
