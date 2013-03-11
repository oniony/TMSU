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

func (db Database) CreateSchema() error {
	var sql string
	var err error

	sql = `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file (
               id INTEGER PRIMARY KEY,
               directory TEXT NOT NULL,
               name TEXT NOT NULL,
               fingerprint TEXT NOT NULL,
               mod_time DATETIME NOT NULL,
               size INTEGER NOT NULL,
               is_dir BOOLEAN NOT NULL,
               CONSTRAINT con_file_path UNIQUE (directory, name)
           )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file_tag (
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
               PRIMARY KEY (file_id, tag_id),
               FOREIGN KEY (file_id) REFERENCES file(id),
               FOREIGN KEY (tag_id) REFERENCES tag(id)
           )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS implication (
                tag_id INTEGER NOT NULL,
                implied_tag_id INTEGER_NOT_NULL,
                PRIMARY KEY (tag_id, implied_tag_id)
           )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	return nil
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
