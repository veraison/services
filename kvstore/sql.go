// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/setrofim/viper"
)

var (
	DefaultTableName = "kvstore"
)

var (
	safeTblNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func isSafeTblName(s string) bool {
	return safeTblNameRe.MatchString(s)
}

type cfg struct {
	TableName      string                 `mapstructure:"tablename"`
	DriverName     string                 `mapstructure:"driver"`
	DataSourceName string                 `mapstructure:"datasource"`
	Other          map[string]interface{} `mapstructure:",remain"`
}

func (o *cfg) Validate() error {
	if !isSafeTblName(o.TableName) {
		return fmt.Errorf("unsafe table name: %q (MUST match %s)",
			o.TableName, safeTblNameRe)
	}

	if o.DriverName == "" {
		return errors.New("\"sql.driver\" directive not found")
	}

	if o.DataSourceName == "" {
		return errors.New("\"sql.datasource\" directive not found")
	}

	if o.Other != nil {
		// Print the textual representation of the map with the "map[]"
		// stripped, resulting a list of "<key>:<value>" entries.
		other := fmt.Sprintf("%v", o.Other)
		return fmt.Errorf(`unexpected "sql" directive(s): %v`, other[4:len(other)-1])
	}

	return nil
}

type SQL struct {
	TableName string
	DB        *sql.DB
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
func (o *SQL) Init(v *viper.Viper) error {
	var cfg cfg

	v.SetDefault("sql.tablename", DefaultTableName)
	if err := v.UnmarshalKey("sql", &cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	o.TableName = cfg.TableName

	db, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	if err != nil {
		return err
	}

	o.DB = db

	return nil
}

func (o *SQL) Close() error {
	return o.DB.Close()
}

func (o SQL) Setup() error {
	if o.DB == nil {
		return errors.New("SQL store uninitialized")
	}

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("CREATE TABLE %s (key text NOT NULL, vals text NOT NULL)", o.TableName)
	_, err := o.DB.Exec(q)

	return err
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
