// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/veraison/services/policy"
)

var (
	addCmd = &cobra.Command{
		Use:      "add [-s STORE] ID FILE",
		Short:    "add a new policy, or update an existing one under the specified ID",
		Args:     cobra.MatchAll(cobra.ExactArgs(2), validateAddArgs),
		RunE:     doAddCommand,
		PreRunE:  initPolicyStore,
		PostRunE: finiPolicyStore,
	}

	shouldUpdate bool
)

func init() {
	addCmd.PersistentFlags().BoolVarP(&shouldUpdate, "update", "u", false,
		"if specfied, the policy will be updated if it already exists")
}

func validateAddArgs(cmd *cobra.Command, args []string) error {
	// note: assumes ExactArgs(2) matched.

	if _, err := policy.PolicyKeyFromString(args[0]); err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	if _, err := os.Stat(args[1]); err != nil {
		return fmt.Errorf("could not stat policy file: %w", err)
	}

	return nil
}

func doAddCommand(cmd *cobra.Command, args []string) error {
	policyID, err := policy.PolicyKeyFromString(args[0])
	if err != nil {
		return err
	}

	policyFile := args[1]

	rulesBytes, err := os.ReadFile(policyFile)
	if err != nil {
		return fmt.Errorf("could not read policy: %w", err)
	}

	addFunc := store.Add
	if shouldUpdate {
		addFunc = store.Update
	}

	policy, err := addFunc(policyID, "default", "opa", string(rulesBytes))
	if err != nil {
		return fmt.Errorf("could not add policy: %w", err)
	}

	log.Printf("Policy %q stored under key %q with UUID %q .\n",
		policyFile, policyID, policy.UUID)

	return nil
}
