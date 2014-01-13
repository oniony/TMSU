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
	"fmt"
	"tmsu/entities"
)

// The number of tags in the database.
func (storage *Storage) TagCount() (uint, error) {
	return storage.Db.TagCount()
}

// The set of tags.
func (storage *Storage) Tags() (entities.Tags, error) {
	return storage.Db.Tags()
}

// Retrieves a specific tag.
func (storage Storage) Tag(id uint) (*entities.Tag, error) {
	return storage.Db.Tag(id)
}

// Retrieves a specific set of tags.
func (storage Storage) TagsByIds(ids []uint) (entities.Tags, error) {
	return storage.Db.TagsByIds(ids)
}

// Retrieves a specific tag.
func (storage Storage) TagByName(name string) (*entities.Tag, error) {
	return storage.Db.TagByName(name)
}

// Retrieves the set of named tags.
func (storage Storage) TagsByNames(names []string) (entities.Tags, error) {
	return storage.Db.TagsByNames(names)
}

// Adds a tag.
func (storage *Storage) AddTag(name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	return storage.Db.InsertTag(name)
}

// Renames a tag.
func (storage Storage) RenameTag(tagId uint, name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	return storage.Db.RenameTag(tagId, name)
}

// Copies a tag.
func (storage Storage) CopyTag(sourceTagId uint, name string) (*entities.Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	tag, err := storage.Db.InsertTag(name)
	if err != nil {
		return nil, fmt.Errorf("could not create tag '%v': %v", name, err)
	}

	err = storage.Db.CopyFileTags(sourceTagId, tag.Id)
	if err != nil {
		return nil, fmt.Errorf("could not copy file tags for tag #%v to tag '%v': %v", sourceTagId, name, err)
	}

	return tag, nil
}

// Deletes a tag.
func (storage Storage) DeleteTag(tagId uint) error {
	err := storage.DeleteFileTagsByTagId(tagId)
	if err != nil {
		return err
	}

	err = storage.Db.DeleteTag(tagId)
	if err != nil {
		return fmt.Errorf("could not delete tag '%v': %v", tagId, err)
	}

	return nil
}

// Retrieves the most popular tags.
func (storage Storage) TopTags(count uint) ([]entities.TagFileCount, error) {
	return storage.Db.TopTags(count)
}

// unexported

func validateTagName(tagName string) error {
	switch tagName {
	case "":
		return errors.New("tag name cannot be empty.")
	case ".", "..":
		return errors.New("tag name cannot be '.' or '..'.") // cannot be used in the VFS
	case "and", "or", "not":
		return errors.New("tag name cannot be a logical operator: 'and', 'or' or 'not'.") // used in query language
	}

	if tagName[0] == '-' {
		return errors.New("tag name cannot start with a minus: '-'.") // used in query language
	}

	for _, ch := range tagName {
		switch ch {
		case '(', ')':
			return errors.New("tag names cannot contain parentheses: '(' or ')'.") // used in query language
		case ',':
			return errors.New("tag names cannot contain comma: ','.") // reserved for tag delimiter
		case '=', '<', '>':
			return errors.New("tag names cannot contain a comparison operator: '=', '<' or '>'.") // reserved for tag values
		case ' ', '\t':
			return errors.New("tag names cannot contain space or tab.") // used as tag delimiter
		case '/':
			return errors.New("tag names cannot contain slash: '/'.") // cannot be used in the VFS
		}
	}

	return nil
}

func containsTagId(items []uint, searchItem uint) bool {
	for _, item := range items {
		if item == searchItem {
			return true
		}
	}

	return false
}
