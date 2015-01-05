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

package storage

import (
	"fmt"
	"tmsu/entities"
)

// The complete set of settings.
func (storage *Storage) Settings() (entities.Settings, error) {
	return storage.Db.Settings()
}

// Retrievs the specified setting.
func (storage *Storage) Setting(name string) (*entities.Setting, error) {
	setting, err := storage.Db.Setting(name)
	if err != nil {
		return nil, err
	}

	// defaults
	if setting == nil {
		switch name {
		case "fingerprintAlgorithm":
			return &entities.Setting{name, "dynamic:SHA256"}, nil
		case "autoCreateTags", "autoCreateValues":
			return &entities.Setting{name, "yes"}, nil
		}
	}

	return setting, nil
}

// Retrieves the specified setting's string value.
func (storage *Storage) SettingAsString(name string) (string, error) {
	setting, err := storage.Setting(name)
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "", fmt.Errorf("no such setting '%v'.", name)
	}

	return setting.Value, nil
}

// Retrieves the specified setting's boolean value.
func (storage *Storage) SettingAsBool(name string) (bool, error) {
	setting, err := storage.Setting(name)
	if err != nil {
		return false, err
	}
	if setting == nil {
		return false, fmt.Errorf("no such setting '%v'.", name)
	}

	switch setting.Value {
	case "yes":
		return true, nil
	case "no":
		return false, nil
	default:
		return false, fmt.Errorf("setting '%v' has an invalid value '%v': expected 'yes' or 'no'.", name, setting.Value)

	}
}
