// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"fmt"

	"github.com/setrofim/viper"
	"github.com/spf13/cobra"
	"github.com/veraison/services/policy"
)

var store *policy.Store

func initPolicyStore(cmd *cobra.Command, args []string) error {
	cfg := viper.Sub("po-store")

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
