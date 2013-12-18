/*
Copyright 2011-2013 Paul Ruane.

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
	"errors"
	"tmsu/entities"
)

// Retrieves a spceific value.
func (storage Storage) Value(id uint) (*entities.Value, error) {
	return storage.Db.Value(id)
}

// Retrieves a specific value by name.
func (storage Storage) ValueByName(name string) (*entities.Value, error) {
	if name == "" {
		return &entities.Value{0, ""}, nil
	}

	return storage.Db.ValueByName(name)
}

// Retrieves the set of values for the specified tag.
func (storage *Storage) ValuesByTagId(tagId uint) (entities.Values, error) {
	return storage.Db.ValuesByTagId(tagId)
}

// Adds a value.
func (storage *Storage) AddValue(name string) (*entities.Value, error) {
	if err := validateValueName(name); err != nil {
		return nil, err
	}

	return storage.Db.InsertValue(name)
}

// Deletes a value.
func (storage Storage) DeleteValue(valueId uint) error {
	return storage.Db.DeleteValue(valueId)
}

// unexported

func validateValueName(valueName string) error {
	if valueName == "" {
		return errors.New("value name cannot be empty.")
	}

	//TODO validate value names

	return nil
}
