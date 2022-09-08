// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/config"
)

func TestSQL_Init_invalid_type_for_store_table(t *testing.T) {
	s := SQL{}

	cfg := config.Store{
		"sql.tablename":  -1,
		"sql.driver":     "sqlite3",
		"sql.datasource": "db=veraison.sql",
	}

	expectedErr := `invalidly specified directive "sql.tablename": want string, got int`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_driver_name(t *testing.T) {
	s := SQL{}

	cfg := config.Store{
		"sql.tablename":  "trustanchor",
		"sql.datasource": "db=veraison-trustanchor.sql",
	}

	expectedErr := `"sql.driver" directive not found`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_bad_tablename(t *testing.T) {
	s := SQL{}

	attemptedInjection := "kvstore ; DROP TABLE another ; SELECT * FROM kvstore"

	cfg := config.Store{
		"sql.tablename":  attemptedInjection,
		"sql.datasource": "db=veraison-trustanchor.sql",
		"sql.driver":     "sqlite3",
	}

	expectedErr := fmt.Sprintf("unsafe table name: %q (MUST match %s)", attemptedInjection, safeTblNameRe)

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Init_missing_datasource_name(t *testing.T) {
	s := SQL{}

	cfg := config.Store{
		"sql.tablename": "trustanchor",
		"sql.driver":    "postgres",
	}

	expectedErr := `"sql.datasource" directive not found`

	err := s.Init(cfg)
	assert.EqualError(t, err, expectedErr)
}

// SQL drivers need to be anonymously imported by the calling code
func TestSQL_Init_db_open_unknown_driver_postgres(t *testing.T) {
	s := SQL{}

	cfg := config.Store{
		"sql.tablename":  "trustanchor",
		"sql.driver":     "postgres",
		"sql.datasource": "db=veraison-trustanchor.sql",
	}

	expectedErr := `sql: unknown driver "postgres" (forgotten import?)`

	err := s.Init(cfg)
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

	s := SQL{TableName: "endorsement", DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	_, err = s.Get(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Get_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	dbErrorString := "a DB error"

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT vals FROM endorsement WHERE key = ?"))
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

	s := SQL{TableName: "endorsement", DB: db}

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT vals FROM endorsement WHERE key = ?"))
	e.WithArgs("ninja")
	e.WillReturnRows(sqlmock.NewRows([]string{"key", "vals"}))

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

	s := SQL{TableName: "endorsement", DB: db}

	rows := sqlmock.NewRows([]string{"vals"})
	rows.AddRow(nil)

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT vals FROM endorsement WHERE key = ?"))
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

	rows := sqlmock.NewRows([]string{"vals"})
	rows.AddRow("[1, 2]")

	e := mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT vals FROM endorsement WHERE key = ?"))
	e.WithArgs("key")
	e.WillReturnRows(rows)

	s := SQL{TableName: "endorsement", DB: db}

	vals, err := s.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"[1, 2]"}, vals)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Set_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_bad_val(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	invalidJSON := ""

	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Set(testKey, invalidJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Set_db_layer_delete_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	dbErrorString := "a DB error"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?")).
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

	s := SQL{TableName: "endorsement", DB: db}

	dbErrorString := "a DB error"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)")).
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

	s := SQL{TableName: "endorsement", DB: db}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?")).
		WithArgs(testKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)")).
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

	s := SQL{TableName: "endorsement", DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Del_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	dbErrorString := "a DB error"

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
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

	s := SQL{TableName: "endorsement", DB: db}

	e := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM endorsement WHERE key = ?"))
	e.WithArgs(testKey)
	e.WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Del(testKey)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSQL_Add_empty_key(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	emptyKey := ""

	expectedErr := `the supplied key is empty`

	err = s.Add(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Add_bad_val(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	invalidJSON := ""

	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Add(testKey, invalidJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestSQL_Add_db_layer_failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := SQL{TableName: "endorsement", DB: db}

	dbErrorString := "a DB error"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)")).
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

	s := SQL{TableName: "endorsement", DB: db}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO endorsement(key, vals) VALUES(?, ?)")).
		WithArgs(testKey, testVal).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Add(testKey, testVal)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
