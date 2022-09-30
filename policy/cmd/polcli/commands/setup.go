// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"github.com/spf13/cobra"
)

var (
	setupCmd = &cobra.Command{
		Use:      "setup [-s STORE]",
		Short:    "one-time setup for a new store.",
		Long:     "Perform a one-time setup of the store. What this entails is backend-dependent (e.g. the sql backend will create the table used by the store.",
		Args:     cobra.NoArgs,
		RunE:     doSetupCommand,
		PreRunE:  initPolicyStore,
		PostRunE: finiPolicyStore,
	}
)

func doSetupCommand(cmd *cobra.Command, args []string) error {
	return store.Setup()
}
