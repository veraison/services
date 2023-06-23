// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/veraison/services/policy"
)

var (
	getCmd = &cobra.Command{
		Use:      "get [-s STORE] [-v VERSION] ID",
		Short:    "get the policy stored under the specified ID",
		Args:     cobra.MatchAll(cobra.ExactArgs(1), validateGetArgs),
		RunE:     doGetCommand,
		PreRunE:  initPolicyStore,
		PostRunE: finiPolicyStore,
	}

	getUUID           string
	getOutputFilePath string
)

func init() {
	getCmd.PersistentFlags().StringVarP(&getUUID, "version", "v", "",
		"get the specified, rather than latest, version")
	getCmd.PersistentFlags().StringVarP(&getOutputFilePath, "output", "o", "",
		"write the policy to the specified file, rather than STDOUT")
}

func validateGetArgs(cmd *cobra.Command, args []string) error {
	// note: assumes ExactArgs(1) matched.

	if _, err := policy.PolicyKeyFromString(args[0]); err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	return nil
}

func doGetCommand(cmd *cobra.Command, args []string) error {
	var policies []*policy.Policy
	var pol *policy.Policy
	var err error

	policyKey, err := policy.PolicyKeyFromString(args[0])
	if err != nil {
		return err
	}

	if getUUID == "" {
		pol, err = store.GetActive(policyKey)
		if err != nil {
			return err
		}
	} else {
		policies, err = store.Get(policyKey)
		if err != nil {
			return err
		}

		found := false
		for _, candidate := range policies {
			if candidate.UUID.String() == getUUID {
				pol = candidate
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("UUID %q for policy %q not found",
				getUUID, policyKey)
		}
	}

	var writer *os.File

	if getOutputFilePath != "" {
		writer, err = os.Create(getOutputFilePath)
		if err != nil {
			return fmt.Errorf("Could not open %q for writing: %w",
				getOutputFilePath, err)
		}
	} else {
		writer = os.Stdout
	}

	if _, err := writer.Write([]byte(pol.Rules)); err != nil {
		return err
	}

	return nil
}
