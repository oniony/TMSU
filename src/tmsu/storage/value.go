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
	"errors"
	"tmsu/entities"
)

// Retrievse the count of values.
func (storage *Storage) ValueCount() (uint, error) {
	return storage.Db.ValueCount()
}

// Retrieves the complete set of values.
func (storage *Storage) Values() (entities.Values, error) {
	return storage.Db.Values()
}

// Retrieves a specific value.
func (storage *Storage) Value(id uint) (*entities.Value, error) {
	return storage.Db.Value(id)
}

// Retrievse the set of unused values.
func (storage *Storage) UnusedValues() (entities.Values, error) {
	return storage.Db.UnusedValues()
}

// Retrieves a specific value by name.
func (storage *Storage) ValueByName(name string) (*entities.Value, error) {
	if name == "" {
		return &entities.Value{0, ""}, nil
	}

	return storage.Db.ValueByName(name)
}

// Retrieves the set of values for the specified tag.
func (storage *Storage) ValuesByTag(tagId uint) (entities.Values, error) {
	return storage.Db.ValuesByTagId(tagId)
}

// Retrieves the set of values with the specified names.
func (storage *Storage) ValuesByNames(names []string) (entities.Values, error) {
	return storage.Db.ValuesByNames(names)
}

// Adds a value.
func (storage *Storage) AddValue(name string) (*entities.Value, error) {
	if err := validateValueName(name); err != nil {
		return nil, err
	}

	return storage.Db.InsertValue(name)
}

// Deletes a value.
func (storage *Storage) DeleteValue(valueId uint) error {
	if err := storage.Db.DeleteValue(valueId); err != nil {
		return err
	}

	return storage.DeleteUnusedValues()
}

// Deletes unused values.
func (storage *Storage) DeleteUnusedValues() error {
	return storage.Db.DeleteUnusedValues()
}

// unexported

func validateValueName(valueName string) error {
	if valueName == "" {
		return errors.New("value name cannot be empty.")
	}

	//TODO validate value names

	return nil
}
