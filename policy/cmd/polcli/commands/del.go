// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/veraison/services/policy"
)

var (
	delCmd = &cobra.Command{
		Use:      "del [-s STORE] ID",
		Short:    "delete the policy with the specified ID",
		Args:     cobra.MatchAll(cobra.ExactArgs(1), validateDelArgs),
		RunE:     doDelCommand,
		PreRunE:  initPolicyStore,
		PostRunE: finiPolicyStore,
	}
)

func validateDelArgs(cmd *cobra.Command, args []string) error {
	// note: assumes ExactArgs(1) matched.

	if _, err := policy.PolicyKeyFromString(args[0]); err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	return nil
}

func doDelCommand(cmd *cobra.Command, args []string) error {
	policyID, err := policy.PolicyKeyFromString(args[0])
	if err != nil {
		return err
	}

	if err := store.Del(policyID); err != nil {
		return fmt.Errorf("could not delete policy: %w", err)
	}

	log.Printf("Policy %q deleted.\n", policyID)

	return nil
}
