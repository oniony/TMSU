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

package storage

import (
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage/database"
)

// The number of tags in the database.
func (storage *Storage) TagCount(tx *Tx) (uint, error) {
	return database.TagCount(tx.tx)
}

// The set of tags.
func (storage *Storage) Tags(tx *Tx) (entities.Tags, error) {
	return database.Tags(tx.tx)
}

// Retrieves a specific tag.
func (storage Storage) Tag(tx *Tx, id entities.TagId) (*entities.Tag, error) {
	return database.Tag(tx.tx, id)
}

// Retrieves a specific set of tags.
func (storage Storage) TagsByIds(tx *Tx, ids entities.TagIds) (entities.Tags, error) {
	return database.TagsByIds(tx.tx, ids)
}

// Retrieves a specific tag.
func (storage Storage) TagByName(tx *Tx, name string) (*entities.Tag, error) {
	return storage.TagByCasedName(tx, name, false)
}

// Retrieves a specific tag with specified case-sensitivity.
func (storage Storage) TagByCasedName(tx *Tx, name string, ignoreCase bool) (*entities.Tag, error) {
	return database.TagByName(tx.tx, name, ignoreCase)
}

// Retrieves the set of named tags.
func (storage Storage) TagsByNames(tx *Tx, names []string) (entities.Tags, error) {
	return storage.TagsByCasedNames(tx, names, false)
}

// Retrieves the set of named tags.
func (storage Storage) TagsByCasedNames(tx *Tx, names []string, ignoreCase bool) (entities.Tags, error) {
	return database.TagsByNames(tx.tx, names, ignoreCase)
}

// Adds a tag.
func (storage *Storage) AddTag(tx *Tx, name string) (*entities.Tag, error) {
	if err := entities.ValidateTagName(name); err != nil {
		return nil, err
	}

	return database.InsertTag(tx.tx, name)
}

// Renames a tag.
func (storage Storage) RenameTag(tx *Tx, tagId entities.TagId, name string) (*entities.Tag, error) {
	if err := entities.ValidateTagName(name); err != nil {
		return nil, err
	}

	return database.RenameTag(tx.tx, tagId, name)
}

// Copies a tag.
func (storage Storage) CopyTag(tx *Tx, sourceTagId entities.TagId, name string) (*entities.Tag, error) {
	if err := entities.ValidateTagName(name); err != nil {
		return nil, err
	}

	tag, err := database.InsertTag(tx.tx, name)
	if err != nil {
		return nil, err
	}

	err = database.CopyFileTags(tx.tx, sourceTagId, tag.Id)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Deletes a tag.
func (storage Storage) DeleteTag(tx *Tx, tagId entities.TagId) error {
	if err := storage.DeleteFileTagsByTagId(tx, tagId); err != nil {
		return err
	}

	if err := storage.DeleteImplicationsByTagId(tx, tagId); err != nil {
		return err
	}

	if err := database.DeleteTag(tx.tx, tagId); err != nil {
		return err
	}

	return nil
}

// Retrieves the tag usage.
func (storage Storage) TagUsage(tx *Tx) ([]entities.TagFileCount, error) {
	return database.TagUsage(tx.tx)
}
