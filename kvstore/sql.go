// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"go.uber.org/zap"
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

type sqlConfig struct {
	TableName      string `mapstructure:"tablename"`
	DriverName     string `mapstructure:"driver"`
	DataSourceName string `mapstructure:"datasource"`
}

func (o *sqlConfig) Validate() error {
	if !isSafeTblName(o.TableName) {
		return fmt.Errorf("unsafe table name: %q (MUST match %s)",
			o.TableName, safeTblNameRe)
	}

	return nil
}

type SQL struct {
	TableName string
	DB        *sql.DB

	logger *zap.SugaredLogger
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
func (o *SQL) Init(v *viper.Viper, logger *zap.SugaredLogger) error {
	o.logger = logger

	cfg := sqlConfig{
		TableName: DefaultTableName,
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v.Sub("sql")); err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	o.TableName = cfg.TableName

	db, err := sql.Open(cfg.DriverName, cfg.DataSourceName)
	if err != nil {
		return err
	}

	o.DB = db
	o.logger.Infow("store opened", "driver", cfg.DriverName,
		"datasource", cfg.DataSourceName, "table", o.TableName)

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
	o.logger.Debugw("create table", "table", o.TableName)
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

func (o SQL) GetKeys() ([]string, error) {
	if o.DB == nil {
		return nil, errors.New("SQL store uninitialized")
	}

	// nolint:gosec
	// o.TableName has been checked by isSafeTblName on init
	q := fmt.Sprintf("SELECT DISTINCT key FROM %s", o.TableName)

	rows, err := o.DB.Query(q)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var keys []string

	for rows.Next() {
		var s sql.NullString

		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		if !s.Valid {
			panic("broken invariant: found key with null string")
		}

		keys = append(keys, s.String)
	}

	return keys, nil
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

	res, err := o.DB.Exec(q, key)
	if err != nil {
		return err
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if numRows == 0 {
		return fmt.Errorf("%w: %q", ErrKeyNotFound, key)
	}

	return nil
}
