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

package storage

import (
	"tmsu/entities"
	"tmsu/storage/database"
)

var defaultSettings = map[string]string{
	"autoCreateTags":                "yes",
	"autoCreateValues":              "yes",
	"fileFingerprintAlgorithm":      "dynamic:SHA256",
	"directoryFingerprintAlgorithm": "none",
}

// The complete set of settings.
func (storage *Storage) Settings(tx *Tx) (entities.Settings, error) {
	settings, err := database.Settings(tx.tx)
	if err != nil {
		return nil, err
	}

	// enrich with defaults
	for name, value := range defaultSettings {
		if !settings.ContainsName(name) {
			settings = append(settings, &entities.Setting{name, value})
		}
	}

	return settings, nil
}

func (storage *Storage) Setting(tx *Tx, name string) (*entities.Setting, error) {
	setting, err := database.Setting(tx.tx, name)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		value, ok := defaultSettings[name]
		if !ok {
			return nil, nil
		}

		setting = &entities.Setting{name, value}
	}

	return setting, nil
}

func (storage *Storage) UpdateSetting(tx *Tx, name, value string) (*entities.Setting, error) {
	return database.UpdateSetting(tx.tx, name, value)
}
