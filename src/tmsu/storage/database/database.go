/*
Copyright 2011-2014 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"os/user"
	"path/filepath"
	"tmsu/common/log"
)

var Path string

type Database struct {
	Path string

	// unexported
	connection  *sql.DB
	transaction *sql.Tx
}

// Opens the database
func Open() (*Database, error) {
	// attempt to create database directory
	dir := filepath.Dir(Path)
	os.MkdirAll(dir, os.ModeDir|0755)

	return OpenAt(Path)
}

// Opens the database at the specified path
func OpenAt(path string) (*Database, error) {
	log.Infof(2, "opening database at '%v'.", path)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("creating database at '%v'.", path)
		} else {
			log.Warnf("could not stat database: %v", err)
		}
	}

	connection, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, DatabaseAccessError{path, err}
	}

	database := &Database{path, connection, nil}

	if err := database.Begin(); err != nil {
		return nil, err
	}

	if err := database.CreateSchema(); err != nil {
		return nil, err
	}

	if err := database.Commit(); err != nil {
		return nil, err
	}

	return database, nil
}

// Executes a SQL query.
func (db *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	if log.Verbosity >= 3 {
		log.Infof(3, "executing update\n"+query)

		for index, arg := range args {
			log.Infof(3, "Arg %v: %v", index, arg)
		}
	}

	var result sql.Result
	var err error

	if db.transaction != nil {
		result, err = db.transaction.Exec(query, args...)
	} else {
		result, err = db.connection.Exec(query, args...)
	}

	if err != nil {
		return nil, DatabaseQueryError{db.Path, query, err}
	}

	return result, nil
}

// Executes a SQL query returning rows.
func (db *Database) ExecQuery(query string, args ...interface{}) (*sql.Rows, error) {
	if log.Verbosity >= 3 {
		log.Infof(3, "executing query\n"+query)

		for index, arg := range args {
			log.Infof(3, "Arg %v: %v", index, arg)
		}
	}

	var rows *sql.Rows
	var err error

	if db.transaction != nil {
		rows, err = db.transaction.Query(query, args...)
	} else {
		rows, err = db.connection.Query(query, args...)
	}

	if err != nil {
		return nil, DatabaseQueryError{db.Path, query, err}
	}

	return rows, nil
}

// Start a transaction
func (db *Database) Begin() error {
	if db.transaction != nil {
		panic("could not begin transaction: there is already an open transaction")
	}

	log.Info(2, "beginning new transaction")

	transaction, err := db.connection.Begin()
	if err != nil {
		return DatabaseTransactionError{db.Path, err}
	}

	db.transaction = transaction

	return nil
}

// Commits the current transaction
func (db *Database) Commit() error {
	if db.transaction == nil {
		return fmt.Errorf("could not commit transaction: there is no open transaciton")
	}

	log.Info(2, "committing transaction")

	if err := db.transaction.Commit(); err != nil {
		return DatabaseTransactionError{db.Path, err}
	}

	db.transaction = nil

	return nil
}

// Rolls back the current transaction
func (db *Database) Rollback() error {
	if db.transaction == nil {
		return fmt.Errorf("could not rollback transaction: there is no open transaciton")
	}

	log.Info(2, "rolling back transaction")

	if err := db.transaction.Rollback(); err != nil {
		return DatabaseTransactionError{db.Path, err}
	}

	db.transaction = nil

	return nil
}

// Closes the database connection
func (db *Database) Close() error {
	log.Info(3, "closing database")

	if err := db.connection.Close(); err != nil {
		return DatabaseAccessError{db.Path, err}
	}

	return nil
}

// unexported

func init() {
	if path := os.Getenv("TMSU_DB"); path != "" {
		log.Info(3, "TMSU_DB=", path)
		Path = path
	} else {
		u, err := user.Current()
		if err != nil {
			panic(fmt.Sprintf("Could not identify current user: %v", err))
		}

		Path = filepath.Join(u.HomeDir, ".tmsu", "default.db")
	}
}

func readCount(rows *sql.Rows) (uint, error) {
	if !rows.Next() {
		return 0, errors.New("Could not get count.")
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
