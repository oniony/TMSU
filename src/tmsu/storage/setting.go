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

package storage

import (
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
		}
	}

	return setting, nil
}
