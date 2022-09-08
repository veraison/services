// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/veraison/services/config"
)

var (
	DefaultTableName = "kvstore"
)

var (
	safeTblNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type SQL struct {
	TableName string
	DB        *sql.DB
}

func isSafeTblName(s string) bool {
	return safeTblNameRe.MatchString(s)
}

// Init initializes the KVStore. The config may contain the following values,
// all of which are optional:
// "sql.tablename" - The name of the table with key-values pairs (defaults to
//                 "kvstore".
// "sql.driver" - The SQL driver to use; see
//                https://github.com/golang/go/wiki/SQLDrivers (defaults to
//                "sqlite3").
// "sql.datasource" -  The name of the data source to use. Valid values are
//                     driver-specific (defaults to "db=veraison.sql".
func (o *SQL) Init(cfg config.Store) error {
	tableName, err := config.GetString(cfg, DirectiveSQLTableName, &DefaultTableName)
	if err != nil {
		return err
	}
	o.TableName = tableName

	if !isSafeTblName(o.TableName) {
		return fmt.Errorf("unsafe table name: %q (MUST match %s)", o.TableName, safeTblNameRe)
	}

	driverName, err := config.GetString(cfg, DirectiveSQLDriverName, nil)
	if err != nil {
		return err
	}

	dataSourceName, err := config.GetString(cfg, DirectiveSQLDataSourceName, nil)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}

	o.DB = db

	return nil
}

func (o *SQL) Close() error {
	return o.DB.Close()
}

func (o SQL) Get(key string) ([]string, error) {
	if o.DB == nil {
		return nil, errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return nil, err
	}

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("SELECT DISTINCT vals FROM %s WHERE key = ?", o.TableName)

	rows, err := o.DB.Query(q, key)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var vals []string

	count := 0
	for rows.Next() {
		count++
		var s sql.NullString

		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		if !s.Valid {
			panic("broken invariant: found val with null string")
		}

		vals = append(vals, s.String)
	}

	if count == 0 {
		return nil, fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}

	return vals, nil
}

func (o SQL) Add(key string, val string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("INSERT INTO %s(key, vals) VALUES(?, ?)", o.TableName)

	_, err := o.DB.Exec(q, key, val)
	if err != nil {
		return err
	}

	return nil
}

func (o SQL) Set(key string, val string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	txn, err := o.DB.Begin()
	if err != nil {
		return err
	}

	defer func() { _ = txn.Rollback() }()

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	delQ := fmt.Sprintf("DELETE FROM %s WHERE key = ?", o.TableName)

	if _, err = o.DB.Exec(delQ, key); err != nil {
		return err
	}

	// o.TableName has been checked by isSafeTblName on init
	// nolint:gosec
	insQ := fmt.Sprintf("INSERT INTO %s(key, vals) VALUES(?, ?)", o.TableName)

	if _, err = o.DB.Exec(insQ, key, val); err != nil {
		return err
	}

	return txn.Commit()
}

func (o SQL) Del(key string) error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return err
	}

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("DELETE FROM %s WHERE key = ?", o.TableName)

	_, err := o.DB.Exec(q, key)
	if err != nil {
		return err
	}

	return nil
}
