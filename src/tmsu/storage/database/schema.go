/*
Copyright 2011-2015 Paul Ruane.

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
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"tmsu/common"
)

// unexported

var latestSchemaVersion = common.Version{0, 5, 0}

func (db *Database) schemaVersion() common.Version {
	sql := `SELECT major, minor, patch
            FROM version`

	var major, minor, patch uint

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return common.Version{}
	}
	defer rows.Close()

	if rows.Next() && rows.Err() == nil {
		rows.Scan(&major, &minor, &patch) // ignore errors
	}

	return common.Version{major, minor, patch}
}

func (db *Database) insertSchemaVersion(version common.Version) error {
	sql := `INSERT INTO version (major, minor, patch)
            VALUES (?, ?, ?)`

	result, err := db.Exec(sql, version.Major, version.Minor, version.Patch)
	if err != nil {
		return fmt.Errorf("could not update schema version: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("version could not be inserted: expected exactly one row to be affected")
	}

	return nil
}

func (db *Database) updateSchemaVersion(version common.Version) error {
	sql := `UPDATE version SET major = ?, minor = ?, patch = ?`

	result, err := db.Exec(sql, version.Major, version.Minor, version.Patch)
	if err != nil {
		return fmt.Errorf("could not update schema version: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("version could not be updated: expected exactly one row to be affected")
	}

	return nil
}

func (db *Database) createSchema() error {
	if err := db.createTagTable(); err != nil {
		return err
	}

	if err := db.createFileTable(); err != nil {
		return err
	}

	if err := db.createValueTable(); err != nil {
		return err
	}

	if err := db.createFileTagTable(); err != nil {
		return err
	}

	if err := db.createImplicationTable(); err != nil {
		return err
	}

	if err := db.createQueryTable(); err != nil {
		return err
	}

	if err := db.createSettingTable(); err != nil {
		return err
	}

	if err := db.createVersionTable(); err != nil {
		return err
	}

	if err := db.insertSchemaVersion(latestSchemaVersion); err != nil {
		return err
	}

	return nil
}

func (db *Database) createTagTable() error {
	sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createFileTable() error {
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

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createValueTable() error {
	sql := `CREATE TABLE IF NOT EXISTS value (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL,
                CONSTRAINT con_value_name UNIQUE (name)
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createFileTagTable() error {
	sql := `CREATE TABLE IF NOT EXISTS file_tag (
                file_id INTEGER NOT NULL,
                tag_id INTEGER NOT NULL,
                value_id INTEGER NOT NULL,
                PRIMARY KEY (file_id, tag_id, value_id),
                FOREIGN KEY (file_id) REFERENCES file(id),
                FOREIGN KEY (tag_id) REFERENCES tag(id)
                FOREIGN KEY (value_id) REFERENCES value(id)
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
           ON file_tag(value_id)`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createImplicationTable() error {
	sql := `CREATE TABLE IF NOT EXISTS implication (
                tag_id INTEGER NOT NULL,
                implied_tag_id INTEGER NOT NULL,
                PRIMARY KEY (tag_id, implied_tag_id)
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createQueryTable() error {
	sql := `CREATE TABLE IF NOT EXISTS query (
                text TEXT PRIMARY KEY
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createSettingTable() error {
	sql := `CREATE TABLE IF NOT EXISTS setting (
                name TEXT PRIMARY KEY,
                value TEXT NOT NULL
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) createVersionTable() error {
	sql := `CREATE TABLE IF NOT EXISTS version (
                major NUMBER NOT NULL,
                minor NUMBER NOT NULL,
                patch NUMBER NOT NULL,
                PRIMARY KEY (major, minor, patch)
            )`

	if _, err := db.Exec(sql); err != nil {
		return err
	}

	return nil
}
