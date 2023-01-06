// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetSubs(t *testing.T) {
	v := viper.New()
	err := v.MergeConfigMap(map[string]interface{}{
		"entry1":    map[string]interface{}{},
		"optional1": map[string]interface{}{},
	})
	require.NoError(t, err)

	subs, err := GetSubs(v, "entry1", "entry2", "*optional1", "*optional2")
	assert.Nil(t, subs)
	assert.ErrorContains(t, err, "missing directives in : entry2")

	v.Set("entry2", map[string]interface{}{})

	subs, err = GetSubs(v, "entry1", "entry2", "*optional1", "*optional2")
	require.NoError(t, err)
	assert.NotNil(t, subs["optional2"])
}

func Test_EnvVars(t *testing.T) {
	os.Setenv("PREFIX_ENTRY1.SUBA", "aye")
	defer os.Unsetenv("PREFIX_ENTRY1.SUBA")

	rawConfig := map[string]interface{}{
		"entry1": map[string]interface{}{
			"subA":  "a",
			"sub-b": "b",
		},
		"entry2": 2,
	}

	cfgPath := filepath.Join(t.TempDir(), "config.json")

	data, err := json.Marshal(rawConfig)
	require.NoError(t, err)

	err = os.WriteFile(cfgPath, data, 0644)
	require.NoError(t, err)

	v := viper.New()
	v.SetEnvPrefix("prefix")
	v.AutomaticEnv()
	v.SetConfigFile(cfgPath)

	err = v.ReadInConfig()
	require.NoError(t, err)
	assert.Equal(t, "aye", v.GetString("entry1.subA"))
}
