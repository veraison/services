// Copyright 2022 Contributors to the Veraison project.
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

	ldr, err := plugin.CreateLoader(cfg, logger)
	require.NoError(t, err)
	defer ldr.Close()

	plugin.RegisterUsing(ldr, "mook", MookRPC)
	err = plugin.DiscoverUsing[IMook](ldr)
	require.NoError(t, err)

	plugin.RegisterUsing(ldr, "ammo", AmmoRPC)
	err = plugin.DiscoverUsing[IAmmo](ldr)
	require.NoError(t, err)

	mediaTypes := ldr.GetRegisteredMediaTypes()
	expected := []string{"blaster", "phaser", "tibanna gas", "plasma"}
	assert.ElementsMatch(t, expected, mediaTypes)

	mook, err := plugin.GetHandleByMediaTypeUsing[IMook](ldr, "blaster")
	require.NoError(t, err)
	assert.Equal(t, `blaster goes "pew, pew"`, mook.Shoot())
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
