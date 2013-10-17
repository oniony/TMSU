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
	_ "github.com/mattn/go-sqlite3"
)

func (db Database) CreateSchema() error {
	var sql string

	sql = `CREATE TABLE IF NOT EXISTS tag (
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

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file_tag (
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
               PRIMARY KEY (file_id, tag_id),
               FOREIGN KEY (file_id) REFERENCES file(id),
               FOREIGN KEY (tag_id) REFERENCES tag(id)
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

	sql = `CREATE TABLE IF NOT EXISTS implication (
               tag_id INTEGER NOT NULL,
               implied_tag_id INTEGER_NOT_NULL,
               PRIMARY KEY (tag_id, implied_tag_id)
           )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS query (
               text TEXT PRIMARY KEY
           )`

	if _, err := db.connection.Exec(sql); err != nil {
		return err
	}

	return nil
}
