// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package commands

import (
	"crypto/md5"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:      "list [-s STORE]",
		Short:    "list policies in the store",
		Args:     cobra.NoArgs,
		RunE:     doListCommand,
		PreRunE:  initPolicyStore,
		PostRunE: finiPolicyStore,
	}

	shouldListAll bool
)

func init() {
	listCmd.PersistentFlags().BoolVarP(&shouldListAll, "all", "a", false,
		"if specfied, all stored versions of policies will be listed")
}

func doListCommand(cmd *cobra.Command, args []string) error {
	listFunc := store.List
	if shouldListAll {
		listFunc = store.ListAllVersions
	}

	policies, err := listFunc()
	if err != nil {
		return fmt.Errorf("could not list policies: %w", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"id", "version", "md5sum"})

	for _, p := range policies {
		uuid := p.UUID.String()
		md5sum := fmt.Sprintf("%x", md5.Sum([]byte(p.Rules)))
		table.Append([]string{p.StoreKey.String(), uuid, md5sum})
	}

	table.Render()

	return nil
}
