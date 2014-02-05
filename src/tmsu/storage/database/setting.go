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
	"database/sql"
	"tmsu/entities"
)

// The complete set of settings.
func (db *Database) Settings() (entities.Settings, error) {
	sql := `SELECT name, value
	        FROM setting`

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readSettings(rows, make(entities.Settings, 0, 10))
}

// Retrieves the specified setting.
func (db *Database) Setting(name string) (*entities.Setting, error) {
	sql := `SELECT name, value
            FROM setting
            WHERE name = ?`

	rows, err := db.ExecQuery(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readSetting(rows)
}

//

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
