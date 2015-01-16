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
	"database/sql"
	"fmt"
	"tmsu/entities"
)

// The complete set of settings.
func (db *Database) Settings() (entities.Settings, error) {
	sql := `SELECT name, value
	        FROM setting`

	rows, err := db.ExecQuery(sql)
	if err != nil {
		return entities.Settings{}, err
	}
	defer rows.Close()

	return readSettings(rows)
}

// unexported

func readSettings(rows *sql.Rows) (entities.Settings, error) {
    settings := entities.Settings{"dynamic:SHA256", "dynamic:sumSizes", true, true}

	for {
	    if !rows.Next() {
	        break
        }
        if rows.Err() != nil {
            return entities.Settings{}, rows.Err()
        }

        var name, value string
        err := rows.Scan(&name, &value)
        if err != nil {
            return settings, err
        }

        switch name {
        case "fingerprintAlgorithm", "fileFingerprintAlgorithm":
            settings.FileFingerprintAlgorithm = value
        case "dirFingerprintAlgorithm":
            settings.DirectoryFingerprintAlgorithm = value
        case "autoCreateTags":
            settings.AutoCreateTags, err = parseBool(value)
            if err != nil {
                return settings, err
            }
        case "autoCreateValues":
            settings.AutoCreateValues, err = parseBool(value)
            if err != nil {
                return settings, err
            }
        }
	}

	return settings, nil
}

func parseBool(text string) (bool, error) {
    switch text {
    case "yes":
        return true, nil
    case "no":
        return false, nil
    default:
        return false, fmt.Errorf("could not parse bool '%v': expected 'yes' or 'no'")
    }
}
