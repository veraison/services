// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_YAMLReader_Read(t *testing.T) {
	reader := NewYAMLReader()

	text := []byte(`
myconfig:
   key1: 1
   key2: seven
   key3:
      part1: 2.0
      part2: true
   `)

	n, err := reader.Read(text)
	require.NoError(t, err)
	assert.Equal(t, len(text), n)

	expectedStores := map[string]Store{
		"myconfig": {
			"key1":       1,
			"key2":       "seven",
			"key3.part1": 2.0,
			"key3.part2": true,
		},
	}

	assert.Equal(t, expectedStores, reader.Stores)

	n, err = reader.Read(text)
	require.EqualError(t, err, `store already exists: "myconfig"`)
	assert.Equal(t, 0, n)

	badText := []byte("\tmyconfig:\n  key1: `")

	n, err = reader.Read(badText)
	require.EqualError(t, err, "yaml: found character that cannot start any token")
	assert.Equal(t, 0, n)
}

func Test_YAMLReader_ReadFile(t *testing.T) {
	reader := NewYAMLReader()

	f, err := os.Open("test/config.yaml")
	require.NoError(t, err)
	fi, err := f.Stat()
	require.NoError(t, err)
	expectedSize := int(fi.Size())

	expectedStores := map[string]Store{
		"plugin": {
			"backend":          "go-plugin",
			"go-plugin.folder": "../plugins/bin/",
		},
		"ta-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "ta-store.sql",
		},
		"en-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "en-store.sql",
		},
		"vts-grpc": {
			"server.addr": "127.0.0.1:50051",
		},
	}

	n, err := reader.ReadFile("test/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, expectedSize, n)
	assert.Equal(t, expectedStores, reader.Stores)

	n, err = reader.ReadFile("test/config.yaml")
	assert.Equal(t, 0, n)
	assert.ErrorContains(t, err, "store already exists")
	assert.Equal(t, expectedStores, reader.Stores)
}

func Test_YAMLReader_GetStores(t *testing.T) {
	reader := NewYAMLReader()

	expectedStores := map[string]Store{
		"plugin": {
			"backend":          "go-plugin",
			"go-plugin.folder": "../plugins/bin/",
		},
		"ta-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "ta-store.sql",
		},
		"en-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "en-store.sql",
		},
		"vts-grpc": {
			"server.addr": "127.0.0.1:50051",
		},
	}

	reader.Stores = map[string]Store{
		"plugin": {
			"backend":          "go-plugin",
			"go-plugin.folder": "../plugins/bin/",
		},
		"ta-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "ta-store.sql",
		},
		"en-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "en-store.sql",
		},
		"vts-grpc": {
			"server.addr": "127.0.0.1:50051",
		},
	}

	readStores := reader.GetStores()
	assert.Equal(t, expectedStores, readStores)

	delete(readStores, "plugin")
	assert.Equal(t, expectedStores, reader.GetStores())

}

func Test_YAMLReader_GetStore(t *testing.T) {
	reader := NewYAMLReader()

	reader.Stores = map[string]Store{
		"plugin": {
			"backend":          "go-plugin",
			"go-plugin.folder": "../plugins/bin/",
		},
		"ta-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "ta-store.sql",
		},
		"en-store": {
			"backend":        "sql",
			"sql.driver":     "sqlite3",
			"sql.datasource": "en-store.sql",
		},
		"vts-grpc": {
			"server.addr": "127.0.0.1:50051",
		},
	}

	store, err := reader.GetStore("plugin")
	assert.NoError(t, err)
	assert.Equal(t, reader.Stores["plugin"], store)

	store, err = reader.GetStore("nope")
	assert.Nil(t, store)
	assert.EqualError(t, err, `store not found: "nope"`)

	store = reader.MustGetStore("nope")
	assert.Equal(t, Store{}, store)
}
