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
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"tmsu/common/log"
)

// Opens the database at the specified path
func OpenAt(path string) (*sql.DB, error) {
	log.Infof(2, "opening database at '%v'.", path)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("creating database at '%v'.", path)

			dir := filepath.Dir(path)
			os.Mkdir(dir, 0755)
		} else {
			log.Warnf("could not stat database: %v", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, DatabaseAccessError{path, err}
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, DatabaseTransactionError{path, err}
	}

	if err := Upgrade(tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, DatabaseTransactionError{path, err}
	}

	return db, nil
}

// unexported

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
