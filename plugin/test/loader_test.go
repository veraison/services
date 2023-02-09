// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/log"

	"github.com/veraison/services/plugin"
)

func TestLoader_discover_and_load(t *testing.T) {
	err := buildPlugins([]string{"trooper", "redshirt", "powercell", "gascartridge"})
	require.NoError(t, err)

	cfg := map[string]interface{}{"dir": "bin"}
	logger := log.Named("test")

	ldr, err := plugin.CreateGoPluginLoader(cfg, logger)
	require.NoError(t, err)
	defer ldr.Close()

	err = plugin.RegisterGoPluginUsing(ldr, "mook", MookRPC)
	require.NoError(t, err)

	err = plugin.DiscoverGoPluginUsing[IMook](ldr)
	require.NoError(t, err)

	err = plugin.RegisterGoPluginUsing(ldr, "ammo", AmmoRPC)
	require.NoError(t, err)

	err = plugin.DiscoverGoPluginUsing[IAmmo](ldr)
	require.NoError(t, err)

	mediaTypes := ldr.GetRegisteredMediaTypes()
	expected := []string{"blaster", "phaser", "tibanna gas", "plasma"}
	assert.ElementsMatch(t, expected, mediaTypes)
}

func buildPlugins(names []string) error {
	for _, name := range names {
		if err := buildPlugin(name); err != nil {
			return err
		}
	}

	return nil
}

func buildPlugin(name string) error {
	cmd := exec.Command("make", name)
	return cmd.Run()
}

func clean() {
	os.RemoveAll("bin")
}

func TestMain(m *testing.M) {
	ret := m.Run()

	clean()

	os.Exit(ret)
}
