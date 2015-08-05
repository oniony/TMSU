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
	"errors"
	"fmt"
	"tmsu/entities"
	"tmsu/storage/database"
	"unicode"
)

// Retrievse the count of values.
func (storage *Storage) ValueCount(tx *Tx) (uint, error) {
	return database.ValueCount(tx.tx)
}

// Retrieves the complete set of values.
func (storage *Storage) Values(tx *Tx) (entities.Values, error) {
	return database.Values(tx.tx)
}

// Retrieves a specific value.
func (storage *Storage) Value(tx *Tx, id entities.ValueId) (*entities.Value, error) {
	return database.Value(tx.tx, id)
}

// Retrieves a specific set of values.
func (storage Storage) ValuesByIds(tx *Tx, ids entities.ValueIds) (entities.Values, error) {
	return database.ValuesByIds(tx.tx, ids)
}

// Retrievse the set of unused values.
func (storage *Storage) UnusedValues(tx *Tx) (entities.Values, error) {
	return database.UnusedValues(tx.tx)
}

// Retrieves a specific value by name.
func (storage *Storage) ValueByName(tx *Tx, name string) (*entities.Value, error) {
	return storage.ValueByCasedName(tx, name, false)
}

// Retrieves a specific value by name.
func (storage *Storage) ValueByCasedName(tx *Tx, name string, ignoreCase bool) (*entities.Value, error) {
	if name == "" {
		return &entities.Value{0, ""}, nil
	}

	return database.ValueByName(tx.tx, name, ignoreCase)
}

// Retrieves the set of values with the specified names.
func (storage *Storage) ValuesByNames(tx *Tx, names []string) (entities.Values, error) {
	return storage.ValuesByCasedNames(tx, names, false)
}

// Retrieves the set of values with the specified names.
func (storage *Storage) ValuesByCasedNames(tx *Tx, names []string, ignoreCase bool) (entities.Values, error) {
	return database.ValuesByNames(tx.tx, names, ignoreCase)
}

// Retrieves the set of values for the specified tag.
func (storage *Storage) ValuesByTag(tx *Tx, tagId entities.TagId) (entities.Values, error) {
	return database.ValuesByTagId(tx.tx, tagId)
}

// Adds a value.
func (storage *Storage) AddValue(tx *Tx, name string) (*entities.Value, error) {
	if err := validateValueName(name); err != nil {
		return nil, err
	}

	return database.InsertValue(tx.tx, name)
}

// Renames a value.
func (storage *Storage) RenameValue(tx *Tx, valueId entities.ValueId, newName string) (*entities.Value, error) {
	if err := validateValueName(newName); err != nil {
		return nil, err
	}

	return database.RenameValue(tx.tx, valueId, newName)
}

// Deletes a value.
func (storage *Storage) DeleteValue(tx *Tx, valueId entities.ValueId) error {
	if err := storage.DeleteFileTagsByValueId(tx, valueId); err != nil {
		return err
	}

	if err := storage.DeleteImplicationsByValueId(tx, valueId); err != nil {
		return err
	}

	if err := database.DeleteValue(tx.tx, valueId); err != nil {
		return err
	}

	return nil
}

// unexported

var validValueChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol}

func validateValueName(valueName string) error {
	switch valueName {
	case "":
		return errors.New("tag value cannot be empty")
	case ".", "..":
		return errors.New("tag value cannot be '.' or '..'") // cannot be used in the VFS
	case "and", "AND", "or", "OR", "not", "NOT":
		return errors.New("tag value cannot be a logical operator: 'and', 'or' or 'not'") // used in query language
	case "eq", "EQ", "ne", "NE", "lt", "LT", "gt", "GT", "le", "LE", "ge", "GE":
		return errors.New("tag value cannot be a comparison operator: 'eq', 'ne', 'lt', 'gt', 'le' or 'ge'") // used in query language
	}

	for _, ch := range valueName {
		switch ch {
		case '/':
			return errors.New("tag value cannot contain slash: '/'") // cannot be used in the VFS
		case '\\':
			return errors.New("tag names cannot contain backslash: '\\'") // cannot be used in the VFS on Windows
		}

		if !unicode.IsOneOf(validValueChars, ch) {
			return fmt.Errorf("tag value cannot contain '%c'", ch)
		}
	}

	return nil
}
