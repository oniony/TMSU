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
	"github.com/oniony/TMSU/common"
	"github.com/oniony/TMSU/common/log"
)

// unexported

func upgrade(tx *sql.Tx) error {
	version := currentSchemaVersion(tx)

	log.Infof(2, "database schema has version %v, latest schema version is %v", version, latestSchemaVersion)

	if version == latestSchemaVersion {
		log.Infof(2, "schema is up to date")
		return nil
	}

	noVersion := schemaVersion{}
	if version == noVersion {
		log.Infof(2, "creating schema")

		if err := createSchema(tx); err != nil {
			return err
		}

		// still need to run upgrade as per 0.5.0 database did not store a version
	}

	log.Infof(2, "upgrading database")

	if version.LessThan(schemaVersion{common.Version{0, 5, 0}, 0}) {
		log.Infof(2, "renaming fingerprint algorithm setting")

		if err := renameFingerprintAlgorithmSetting(tx); err != nil {
			return err
		}
	}
	if version.LessThan(schemaVersion{common.Version{0, 6, 0}, 0}) {
		log.Infof(2, "recreating implication table")

		if err := recreateImplicationTable(tx); err != nil {
			return err
		}
	}
	if version.LessThan(schemaVersion{common.Version{0, 7, 0}, 0}) {
		log.Infof(2, "updating fingerprint algorithms")

		if err := updateFingerprintAlgorithms(tx); err != nil {
			return err
		}
	}
	if version.LessThan(schemaVersion{common.Version{0, 7, 0}, 1}) {
		log.Infof(2, "recreating version table")

		if err := recreateVersionTable(tx); err != nil {
			return err
		}
	}

	log.Infof(2, "updating schema version")
	if err := updateSchemaVersion(tx, latestSchemaVersion); err != nil {
		return err
	}

	return nil
}

func renameFingerprintAlgorithmSetting(tx *sql.Tx) error {
	if _, err := tx.Exec(`
UPDATE setting
SET name = 'fileFingerprintAlgorithm'
WHERE name = 'fingerprintAlgorithm'`); err != nil {
		return err
	}

	return nil
}

func recreateImplicationTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`
ALTER TABLE implication
RENAME TO implication_old`); err != nil {
		return err
	}

	if err := createImplicationTable(tx); err != nil {
		return err
	}

	if _, err := tx.Exec(`
INSERT INTO implication
SELECT tag_id, 0, implied_tag_id, 0
FROM implication_old`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
DROP TABLE implication_old`); err != nil {
		return err
	}

	return nil
}

func updateFingerprintAlgorithms(tx *sql.Tx) error {
	rows, err := tx.Query(`
SELECT value
FROM setting
WHERE name = 'fileFingerprintAlgorithm'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var fileFingerprintAlg string
	if rows.Next() && rows.Err() == nil {
		rows.Scan(&fileFingerprintAlg) // ignore errors
	}

	switch fileFingerprintAlg {
	case "symlinkTargetName":
		if _, err := tx.Exec(`
INSERT INTO setting (name, value)
VALUES ("symlinkFingerprintAlgorithm", "targetName")`); err != nil {
			return err
		}
		if _, err := tx.Exec(`
DELETE FROM setting WHERE name = 'fileFingerprintAlgorithm';`); err != nil {
			return err
		}
	case "symlinkTargetNameNoExt":
		if _, err := tx.Exec(`
INSERT INTO setting (name, value)
VALUES ("symlinkFingerprintAlgorithm", "targetNameNoExt")`); err != nil {
			return err
		}
		if _, err := tx.Exec(`
DELETE FROM setting WHERE name = 'fileFingerprintAlgorithm';`); err != nil {
			return err
		}
	}

	return nil
}

func recreateVersionTable(tx *sql.Tx) error {
	if _, err := tx.Exec("DROP TABLE version;"); err != nil {
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
