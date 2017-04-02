// Copyright 2011-2017 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3" // initialised Sqlite3
	"github.com/oniony/TMSU/common/log"
	"os"
)

type Database struct {
	db *sql.DB
}

func CreateAt(path string) error {
	log.Infof(2, "creating database at '%v'.", path)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return DatabaseAccessError{path, err}
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return DatabaseTransactionError{path, err}
	}

	if err := upgrade(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return DatabaseTransactionError{path, err}
	}

	return nil
}

func OpenAt(path string) (*Database, error) {
	log.Infof(2, "opening database at '%v'.", path)

	_, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, DatabaseNotFoundError{path}
		default:
			return nil, DatabaseAccessError{path, err}
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, DatabaseAccessError{path, err}
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, DatabaseTransactionError{path, err}
	}

	if err := upgrade(tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, DatabaseTransactionError{path, err}
	}

	return &Database{db}, nil
}

func (database *Database) Close() error {
	return database.db.Close()
}

func (database *Database) Begin() (*Tx, error) {
	tx, err := database.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{tx}, nil
}

type Tx struct {
	tx *sql.Tx
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	log.Info(3, query)
	log.Infof(3, "params: %v", args)

	return tx.tx.Exec(query, args...)
}

func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	log.Info(3, query)
	log.Infof(3, "params: %v", args)

	return tx.tx.Query(query, args...)
}

func (tx *Tx) Commit() error {
	log.Info(2, "committing transaction")

	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	log.Info(2, "rolling back transaction")

	return tx.tx.Rollback()
}

// unexported

func readCount(rows *sql.Rows) (uint, error) {
	if !rows.Next() {
		return 0, errors.New("could not get count")
	}
	if rows.Err() != nil {
		return 0, rows.Err()
	}

	var count uint
	err := rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func collationFor(ignoreCase bool) string {
	if ignoreCase {
		return "COLLATE NOCASE"
	}

	return ""
}
