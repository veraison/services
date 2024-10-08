// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/log"
)

func TestSQL_Init_invalid_type_for_store_table(t *testing.T) {
	s := SQL{}

	cfg := viper.New()
	cfg.Set("Sql.tablename", -1)
	cfg.Set("sql.driver", "sqlite3")
	cfg.Set("sql.datasource", "db=veraison.sql")

	expectedErr := `sql: unsafe table name: "-1" (MUST match ^[a-zA-Z0-9_]+$)`

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_driver_name(t *testing.T) {
	s := SQL{}

	cfg := viper.New()
	cfg.Set("sql.tablename", "trustanchor")
	cfg.Set("sql.datasource", "db=veraison-trustanchor.sql")

	expectedErr := "sql: directives not found: driver"

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_bad_tablename(t *testing.T) {
	s := SQL{}

	attemptedInjection := "kvstore ; DROP TABLE another ; SELECT * FROM kvstore"

	cfg := viper.New()
	cfg.Set("sql.tablename", attemptedInjection)
	cfg.Set("sql.datasource", "db=veraison-trustanchor.sql")
	cfg.Set("sql.driver", "sqlite3")

	expectedErr := fmt.Sprintf("sql: unsafe table name: %q (MUST match %s)", attemptedInjection, safeTblNameRe)

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_datasource_name(t *testing.T) {
	s := SQL{}

	cfg := viper.New()
	cfg.Set("sql.tablename", "trustanchor")
	cfg.Set("sql.driver", "postgres")

	expectedErr := "sql: directives not found: datasource"

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_extra_params(t *testing.T) {
	s := SQL{}

	cfg := viper.New()
	cfg.Set("sql.tablename", "trustanchor")
	cfg.Set("sql.driver", "sqlite3")
	cfg.Set("sql.datasource", "db=veraison-trustanchor.sql")
	cfg.Set("sql.unexpected", "foo")

	expectedErr := "sql: unexpected directives: unexpected"

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

// SQL drivers need to be anonymously imported by the calling code
func TestSQL_Init_db_open_unknown_driver_postgres(t *testing.T) {
	s := SQL{}

	cfg := viper.New()
	cfg.Set("sql.tablename", "trustanchor")
	cfg.Set("sql.driver", "postgres")
	cfg.Set("sql.datasource", "db=veraison-trustanchor.sql")

	expectedErr := `sql: unknown driver "postgres" (forgotten import?)`

	err := s.Init(cfg, log.Named("test"))
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_Get_Del_with_uninitialised_store(t *testing.T) {
	s := SQL{}

	expectedErr := `SQL store uninitialized`

	err := s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Get_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder: sq.Question}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	_, err = s.Get(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Get_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder: sq.Question}

	dbErrorString := "a DB error"

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT kv_val FROM endorsement WHERE kv_key = ?"))
	e.WithArgs("key")
	e.WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	_, err = s.Get("key")
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Get_key_not_found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT kv_val FROM endorsement WHERE kv_key = ?"))
	e.WithArgs("ninja")
	e.WillReturnRows(sqlmock.NewRows([]string{"kv_key", "kv_val"}))

	expectedErr := "key not found: \"ninja\""

	_, err = s.Get("ninja")
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Get_broken_invariant_null_val_panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	rows := sqlmock.NewRows([]string{"kv_val"})
	rows.AddRow(nil)

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT kv_val FROM endorsement WHERE kv_key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	assert.Panics(t, func() { _, _ = s.Get("key") })

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Get_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"kv_val"})
	rows.AddRow("[1, 2]")

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT kv_val FROM endorsement WHERE kv_key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	vals, err := s.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"[1, 2]"}, vals)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_GetKeys_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"key"})
	rows.AddRow("k1")
	rows.AddRow("k2")

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT kv_key FROM endorsement"))
	e.WillReturnRows(rows)

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	keys, err := s.GetKeys()
	assert.NoError(t, err)
	assert.Equal(t, []string{"k1", "k2"}, keys)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Set_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_bad_val(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	invalidJSON := ""

	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Set(testKey, invalidJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_db_layer_delete_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	dbErrorString := "a DB error"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?")).
		WillReturnError(errors.New(dbErrorString))
	mock.ExpectRollback()

	expectedErr := dbErrorString

	err = s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Set_db_layer_insert_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	dbErrorString := "a DB error"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement (kv_key,kv_val) VALUES (?,?)")).
		WillReturnError(errors.New(dbErrorString))
	mock.ExpectRollback()

	expectedErr := dbErrorString

	err = s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
func TestSQL_Set_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?")).
		WithArgs(testKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement (kv_key,kv_val) VALUES (?,?)")).
		WithArgs(testKey, testVal).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = s.Set(testKey, testVal)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Del_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Del_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	dbErrorString := "a DB error"

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?"))
	e.WithArgs(testKey)
	e.WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Del_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?"))
	e.WithArgs(testKey)
	e.WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Del(testKey)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Del_key_not_found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE kv_key = ?"))
	e.WithArgs(testKey)
	e.WillReturnResult(sqlmock.NewResult(1, 0))

	expectedErr := fmt.Sprintf("key not found: %q", testKey)

	err = s.Del(testKey)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Add_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Add(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Add_bad_val(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	invalidJSON := ""

	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Add(testKey, invalidJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Add_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	dbErrorString := "a DB error"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement (kv_key,kv_val) VALUES (?,?)")).
		WithArgs(testKey, testVal).
		WillReturnError(errors.New(dbErrorString))

	expectedErr := dbErrorString

	err = s.Add(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Add_ok(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db, Placeholder:sq.Question}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement (kv_key,kv_val) VALUES (?,?)")).
		WithArgs(testKey, testVal).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Add(testKey, testVal)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Setup(t *testing.T) {
	storeFile := path.Join(t.TempDir(), "store.db")

	cfg := viper.New()
	cfg.Set("sql.driver", "sqlite3")
	cfg.Set("sql.datasource", fmt.Sprintf("file:%s", storeFile))
	cfg.Set("sql.tablename", "test")

	s := SQL{}
	err := s.Init(cfg, log.Named("test"))
	require.NoError(t, err)
	defer s.Close()

	err = s.Setup()
	assert.NoError(t, err)

	err = s.Setup()
	assert.ErrorContains(t, err, "table test already exists")
}
