// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/log"
)

func TestKVStore_New_nil_config(t *testing.T) {
	m, err := New(nil, log.Named("test"))

	expectedErr := `nil configuration`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_missing_backend_directive(t *testing.T) {
	cfg := viper.New()

	m, err := New(cfg, log.Named("test"))

	expectedErr := "directives not found: Backend"

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_unsupported_backend(t *testing.T) {
	cfg := viper.New()
	cfg.Set("backend", "xyz")

	m, err := New(cfg, log.Named("test"))

	expectedErr := `backend "xyz" is not supported`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_unexpected_directives(t *testing.T) {
	cfg := viper.New()

	cfg.Set("backend", "sql")
	cfg.Set("sql.driver", "sqlite3")
	cfg.Set("slq.tablename", "endorsement") // typo in backend name
	cfg.Set("value-type", "string")

	m, err := New(cfg, log.Named("test"))
	expectedErr := `unexpected directives: slq, value-type`

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}

func TestKVStore_New_memory_backend_ok(t *testing.T) {
	cfg := viper.New()
	cfg.Set("backend", "memory")

	m, err := New(cfg, log.Named("test"))

	assert.NoError(t, err)
	assert.IsType(t, &Memory{}, m)
}

func TestKVStore_New_SQL_backend_failed_init(t *testing.T) {
	cfg := viper.New()
	cfg.Set("backend", "sql")
	cfg.Set("sql.tablename", "endorsement")
	cfg.Set("sql.datasource", "db.sql")
	// no sql.driver

	m, err := New(cfg, log.Named("test"))

	expectedErr := "sql: directives not found: driver"

	assert.EqualError(t, err, expectedErr)
	assert.Nil(t, m)
}
