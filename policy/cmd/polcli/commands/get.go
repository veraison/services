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

	getVersion        int32
	getOutputFilePath string
)

func init() {
	getCmd.PersistentFlags().Int32VarP(&getVersion, "version", "v", 0,
		"get the specified, rather than latest, version")
	getCmd.PersistentFlags().StringVarP(&getOutputFilePath, "output", "o", "",
		"write the policy to the specified file, rather than STDOUT")
}

func validateGetArgs(cmd *cobra.Command, args []string) error {
	// note: assumes ExactArgs(1) matched.

	if err := policy.ValidateID(args[0]); err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	return nil
}

func doGetCommand(cmd *cobra.Command, args []string) error {
	var policies []policy.Policy
	var policy policy.Policy
	var err error

	policyID := args[0]

	if getVersion == 0 {
		policy, err = store.GetLatest(policyID)
		if err != nil {
			return err
		}
	} else {
		policies, err = store.Get(policyID)
		if err != nil {
			return err
		}

		for _, candidate := range policies {
			if candidate.Version == getVersion {
				policy = candidate
				break
			}
		}

		if policy.Version == 0 {
			return fmt.Errorf("version %d for policy %q not found",
				getVersion, policyID)
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

	if _, err := writer.Write([]byte(policy.Rules)); err != nil {
		return err
	}

	return nil
}
