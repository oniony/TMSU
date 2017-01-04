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
	"github.com/oniony/TMSU/entities"
)

// The complete set of settings.
func Settings(tx *Tx) (entities.Settings, error) {
	sql := `
SELECT name, value
FROM setting`

	rows, err := tx.Query(sql)
	if err != nil {
		return entities.Settings{}, err
	}
	defer rows.Close()

	settings, err := readSettings(rows, make(entities.Settings, 0, 10))
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func Setting(tx *Tx, name string) (*entities.Setting, error) {
	sql := `
SELECT name, value
FROM setting
WHERE name = ?`

	rows, err := tx.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	setting, err := readSetting(rows)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func UpdateSetting(tx *Tx, name, value string) (*entities.Setting, error) {
	sql := `
INSERT OR REPLACE INTO setting (name, value)
VALUES (?, ?)`

	result, err := tx.Exec(sql, name, value)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, NoSuchSettingError{name}
	}
	if rowsAffected > 1 {
		panic("expected exactly one row to be affected")
	}

	return &entities.Setting{name, value}, nil
}

// unexported

func readSetting(rows *sql.Rows) (*entities.Setting, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var name, value string
	err := rows.Scan(&name, &value)
	if err != nil {
		return nil, err
	}

	return &entities.Setting{name, value}, nil
}

func readSettings(rows *sql.Rows, settings entities.Settings) (entities.Settings, error) {
	for {
		setting, err := readSetting(rows)
		if err != nil {
			return nil, err
		}
		if setting == nil {
			break
		}

		settings = append(settings, setting)
	}

	return settings, nil
}
