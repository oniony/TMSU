/*
Copyright 2011-2012 Paul Ruane.

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
	config, err := common.GetSelectedDatabaseConfig()
	if err != nil {
		return nil, err
	}
	if config == nil {
		config, err = common.GetDefaultDatabaseConfig()
		if err != nil {
			return nil, errors.New("Could not retrieve default database configuration: " + err.Error())
		}

		// attempt to create default database directory
		dir := filepath.Dir(config.DatabasePath)
		os.MkdirAll(dir, os.ModeDir|0755)
	}

	return OpenAt(config.DatabasePath)
}

func OpenAt(path string) (*Database, error) {
	connection, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.New("Could not open database: " + err.Error())
	}

	database := Database{connection}

	err = database.CreateSchema()
	if err != nil {
		return nil, errors.New("Could not create database schema: " + err.Error())
	}

	return &database, nil
}

func (db *Database) Close() error {
	return db.connection.Close()
}

func (db Database) CreateSchema() error {
	sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	_, err := db.connection.Exec(sql)
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

	sql = `CREATE INDEX IF NOT EXISTS idx_file_path
           ON file(directory, name)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file_tag (
               id INTEGER PRIMARY KEY,
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
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

	return nil
}

func (db Database) contains(list []string, str string) bool {
	for _, current := range list {
		if current == str {
			return true
		}
	}

	return false
}
