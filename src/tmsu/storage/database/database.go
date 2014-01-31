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
		return nil, fmt.Errorf("could not open database: %v", err)
	}

	transaction, err := connection.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not begin transaciton: %v", err)
	}

	database := &Database{connection, transaction}

	if err := database.CreateSchema(); err != nil {
		return nil, errors.New("could not create database schema: " + err.Error())
	}

	return database, nil
}

// Executes a SQL query.
func (db *Database) Exec(sql string, args ...interface{}) (sql.Result, error) {
	if log.Verbosity >= 3 {
		log.Infof(3, "executing update\n"+sql)

		for index, arg := range args {
			log.Info(3, "Arg %v = %v", index, arg)
		}
	}

	return db.transaction.Exec(sql, args...)
}

// Executes a SQL query returning rows.
func (db *Database) ExecQuery(sql string, args ...interface{}) (*sql.Rows, error) {
	if log.Verbosity >= 3 {
		log.Infof(3, "executing query\n"+sql)

		for index, arg := range args {
			log.Info(3, "Arg %v = %v", index, arg)
		}
	}

	return db.transaction.Query(sql, args...)
}

// Commits the current transaction
func (db *Database) Commit() error {
	log.Info(2, "committing transaction")

	if err := db.transaction.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	log.Info(2, "beginning new transaction")

	transaction, err := db.connection.Begin()
	if err != nil {
		return fmt.Errorf("could not begin new transaction: %v", err)
	}

	db.transaction = transaction

	return nil
}

// Closes the database connection
func (db *Database) Close() error {
	log.Info(3, "closing database")

	return db.connection.Close()
}

// unexported

func init() {
	if path := os.Getenv("TMSU_DB"); path != "" {
		log.Info(3, "TMSU_DB=", path)
		Path = path
	} else {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("could not identify current user: %v", err)
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
