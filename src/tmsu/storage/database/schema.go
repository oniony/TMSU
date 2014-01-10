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
	_ "github.com/mattn/go-sqlite3"
)

func (db Database) CreateSchema() error {
	if err := db.CreateTagTable(); err != nil {
		return err
	}

	if err := db.CreateFileTable(); err != nil {
		return err
	}

	if err := db.CreateValueTable(); err != nil {
		return err
	}

	if err := db.CreateFileTagTable(); err != nil {
		return err
	}

	if err := db.CreateQueryTable(); err != nil {
		return err
	}

	return nil
}

func (db Database) CreateTagTable() error {
	sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db Database) CreateFileTable() error {
	sql := `CREATE TABLE IF NOT EXISTS file (
                id INTEGER PRIMARY KEY,
                directory TEXT NOT NULL,
                name TEXT NOT NULL,
                fingerprint TEXT NOT NULL,
                mod_time DATETIME NOT NULL,
                size INTEGER NOT NULL,
                is_dir BOOLEAN NOT NULL,
                CONSTRAINT con_file_path UNIQUE (directory, name)
            )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db Database) CreateValueTable() error {
	sql := `CREATE TABLE IF NOT EXISTS value (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL,
                CONSTRAINT con_value_name UNIQUE (name)
            )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db Database) CreateFileTagTable() error {
	sql := `CREATE TABLE IF NOT EXISTS file_tag (
                file_id INTEGER NOT NULL,
                tag_id INTEGER NOT NULL,
                value_id INTEGER NOT NULL,
                PRIMARY KEY (file_id, tag_id, value_id),
                FOREIGN KEY (file_id) REFERENCES file(id),
                FOREIGN KEY (tag_id) REFERENCES tag(id)
                FOREIGN KEY (value_id) REFERENCES value(id)
            )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
           ON file_tag(value_id)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db Database) CreateQueryTable() error {
	sql := `CREATE TABLE IF NOT EXISTS query (
                text TEXT PRIMARY KEY
            )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}
