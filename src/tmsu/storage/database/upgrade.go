// Copyright 2011-2015 Paul Ruane.

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
	_ "github.com/mattn/go-sqlite3"
	"tmsu/common"
	"tmsu/common/log"
)

func Upgrade(tx *sql.Tx) error {
	version := schemaVersion(tx)

	log.Infof(2, "database schema has version %v, latest schema version is %v", version, latestSchemaVersion)

	if version == latestSchemaVersion {
		return nil
	}

	noVersion := common.Version{}
	if version == noVersion {
		log.Infof(2, "creating schema")

		if err := createSchema(tx); err != nil {
			return err
		}

		// still need to run upgrade as per 0.5.0 database did not store a version
	}

	log.Infof(2, "upgrading database")

	if version.LessThan(common.Version{0, 5, 0}) {
		if err := renameFingerprintAlgorithmSetting(tx); err != nil {
			return err
		}
	}

	if err := updateSchemaVersion(tx, latestSchemaVersion); err != nil {
		return err
	}

	return nil
}

// unexported

func renameFingerprintAlgorithmSetting(tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE setting
                          SET name = 'fileFingerprintAlgorithm'
                          WHERE name = 'fingerprintAlgorithm'`)

	if err != nil {
		return fmt.Errorf("could not upgrade database: %v", err)
	}

	return nil
}
