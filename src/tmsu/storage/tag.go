/*
Copyright 2011-2012 Paul Ruane.

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
	"tmsu/storage/database"
)

// The number of tags in the database.
func (storage *Storage) TagCount() (uint, error) {
	return storage.Db.TagCount()
}

// The set of tags.
func (storage *Storage) Tags() (database.Tags, error) {
	return storage.Db.Tags()
}

// Retrieves a specific tag.
func (storage Storage) TagByName(name string) (*database.Tag, error) {
	return storage.Db.TagByName(name)
}

// The set of further tags for which there are tagged files given
// a particular set of tags.
func (storage Storage) TagsForTags(tagIds []uint) (database.Tags, error) {
	files, err := storage.FilesWithTags(tagIds, []uint{}, false)
	if err != nil {
		return nil, err
	}

	furtherTags := make(database.Tags, 0, 10)
	for _, file := range files {
		tags, err := storage.TagsByFileId(file.Id, false)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags {
			if !containsTagId(tagIds, tag.Id) && !containsTag(furtherTags, tag) {
				furtherTags = append(furtherTags, tag)
			}
		}
	}

	return furtherTags, nil
}

// Adds a tag.
func (storage *Storage) AddTag(name string) (*database.Tag, error) {
	return storage.Db.InsertTag(name)
}

// Renames a tag.
func (storage Storage) RenameTag(tagId uint, name string) (*database.Tag, error) {
	return storage.Db.RenameTag(tagId, name)
}

// Copies a tag.
func (storage Storage) CopyTag(sourceTagId uint, name string) (*database.Tag, error) {
	tag, err := storage.Db.InsertTag(name)
	if err != nil {
		return nil, err
	}

	err = storage.Db.CopyFileTags(sourceTagId, tag.Id)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Deletes a tag.
func (storage Storage) DeleteTag(tagId uint) error {
	return storage.Db.DeleteTag(tagId)
}

// 

func containsTagId(items []uint, searchItem uint) bool {
	for _, item := range items {
		if item == searchItem {
			return true
		}
	}

	return false
}

func containsTag(tags database.Tags, searchTag *database.Tag) bool {
	for _, tag := range tags {
		if tag.Id == searchTag.Id {
			return true
		}
	}

	return false
}
