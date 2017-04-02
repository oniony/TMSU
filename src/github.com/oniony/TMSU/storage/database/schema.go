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
	"fmt"
	"github.com/oniony/TMSU/common"
)

// unexported

var latestSchemaVersion = schemaVersion{common.Version{0, 7, 0}, 1}

func currentSchemaVersion(tx *sql.Tx) schemaVersion {
	sql := `
SELECT *
FROM version`

	var major, minor, patch, revision uint

	rows, err := tx.Query(sql)
	if err != nil {
		return schemaVersion{}
	}
	defer rows.Close()

	if rows.Next() && rows.Err() == nil {
		err := rows.Scan(&major, &minor, &patch, &revision)
		if err != nil {
			// might be prior to schema 0.7.0-1 where there was no revision column
			rows.Scan(&major, &minor, &patch) // ignore errors
		}
	}

	return schemaVersion{common.Version{major, minor, patch}, revision}
}

func insertSchemaVersion(tx *sql.Tx, version schemaVersion) error {
	sql := `
INSERT INTO version (major, minor, patch, revision)
VALUES (?, ?, ?, ?)`

	result, err := tx.Exec(sql, version.Major, version.Minor, version.Patch, version.Revision)
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

func updateSchemaVersion(tx *sql.Tx, version schemaVersion) error {
	sql := `
UPDATE version SET major = ?, minor = ?, patch = ?, revision = ?`

	result, err := tx.Exec(sql, version.Major, version.Minor, version.Patch, version.Revision)
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

func createSchema(tx *sql.Tx) error {
	if err := createTagTable(tx); err != nil {
		return err
	}

	if err := createFileTable(tx); err != nil {
		return err
	}

	if err := createValueTable(tx); err != nil {
		return err
	}

	if err := createFileTagTable(tx); err != nil {
		return err
	}

	if err := createImplicationTable(tx); err != nil {
		return err
	}

	if err := createQueryTable(tx); err != nil {
		return err
	}

	if err := createSettingTable(tx); err != nil {
		return err
	}

	if err := createVersionTable(tx); err != nil {
		return err
	}

	if err := insertSchemaVersion(tx, latestSchemaVersion); err != nil {
		return err
	}

	return nil
}

func createTagTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	sql = `
CREATE INDEX IF NOT EXISTS idx_tag_name
ON tag(name)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createFileTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS file (
    id INTEGER PRIMARY KEY,
    directory TEXT NOT NULL,
    name TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    mod_time DATETIME NOT NULL,
    size INTEGER NOT NULL,
    is_dir BOOLEAN NOT NULL,
    CONSTRAINT con_file_path UNIQUE (directory, name)
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	sql = `
CREATE INDEX IF NOT EXISTS idx_file_fingerprint
ON file(fingerprint)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createValueTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS value (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT con_value_name UNIQUE (name)
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createFileTagTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id, value_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
    FOREIGN KEY (value_id) REFERENCES value(id)
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	sql = `
CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
ON file_tag(file_id)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	sql = `
CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
ON file_tag(tag_id)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	sql = `
CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
ON file_tag(value_id)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createImplicationTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS implication (
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    implied_tag_id INTEGER NOT NULL,
    implied_value_id INTEGER NOT NULL,
    PRIMARY KEY (tag_id, value_id, implied_tag_id, implied_value_id)
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createQueryTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS query (
    text TEXT PRIMARY KEY
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createSettingTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS setting (
    name TEXT PRIMARY KEY,
    value TEXT NOT NULL
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}

func createVersionTable(tx *sql.Tx) error {
	sql := `
CREATE TABLE IF NOT EXISTS version (
    major NUMBER NOT NULL,
    minor NUMBER NOT NULL,
    patch NUMBER NOT NULL,
    revision NUMBER NOT NULL,
    PRIMARY KEY (major, minor, patch, revision)
)`

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	return nil
}
