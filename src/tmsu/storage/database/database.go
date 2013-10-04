/*
Copyright 2011-2013 Paul Ruane.

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
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"tmsu/common"
	"tmsu/log"
)

type Database struct {
	connection *sql.DB
}

func Open() (*Database, error) {
	databasePath, err := common.GetDatabasePath()
	if err != nil {
		return nil, err
	}

	// attempt to create database directory
	dir := filepath.Dir(databasePath)
	os.MkdirAll(dir, os.ModeDir|0755)

	return OpenAt(databasePath)
}

func OpenAt(path string) (*Database, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("Creating database at '%v'.", path)
		} else {
			log.Warnf("Could not stat database: %v", err)
		}
	}

	connection, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.New("could not open database: " + err.Error())
	}

	database := Database{connection}

	err = database.CreateSchema()
	if err != nil {
		return nil, errors.New("could not create database schema: " + err.Error())
	}

	return &database, nil
}

func (db *Database) Close() error {
	return db.connection.Close()
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
