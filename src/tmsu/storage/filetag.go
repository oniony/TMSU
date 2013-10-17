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
	"tmsu/storage/entities"
)

// Determines whether the specified file has the specified tag applied.
func (storage *Storage) FileTagExists(fileId, tagId uint) (bool, error) {
	return storage.Db.FileTagExists(fileId, tagId)
}

// Retrieves the total count of file tags in the database.
func (storage *Storage) FileTagCount() (uint, error) {
	return storage.Db.FileTagCount()
}

// Retrieves the complete set of file tags.
func (storage *Storage) FileTags() (entities.FileTags, error) {
	return storage.Db.FileTags()
}

// Retrieves the count of file tags for the specified file.
func (storage *Storage) FileTagCountByFileId(fileId uint) (uint, error) {
	return storage.Db.FileTagCountByFileId(fileId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint) (entities.FileTags, error) {
	return storage.Db.FileTagsByTagId(tagId)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId uint) (entities.FileTags, error) {
	return storage.Db.FileTagsByFileId(fileId)
}

// Adds an file tag.
func (storage *Storage) AddFileTag(fileId, tagId uint) (*entities.FileTag, error) {
	return storage.Db.AddFileTag(fileId, tagId)
}

// Adds a set of file tags.
func (storage *Storage) AddFileTags(fileId uint, tagIds []uint) error {
	if len(tagIds) == 0 {
		return nil
	}

	return storage.Db.AddFileTags(fileId, tagIds)
}

// Remove file tag.
func (storage *Storage) RemoveFileTag(fileId, tagId uint) error {
	return storage.Db.DeleteFileTag(fileId, tagId)
}

// Removes all of the file tags for the specified file.
func (storage *Storage) RemoveFileTagsByFileId(fileId uint) error {
	return storage.Db.DeleteFileTagsByFileId(fileId)
}

// Removes all of the file tags for the specified tag.
func (storage *Storage) RemoveFileTagsByTagId(tagId uint) error {
	return storage.Db.DeleteFileTagsByTagId(tagId)
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId, destTagId uint) error {
	return storage.Db.CopyFileTags(sourceTagId, destTagId)
}

// helpers

func contains(files entities.Files, searchFile *entities.File) bool {
	for _, file := range files {
		if file.Path() == searchFile.Path() {
			return true
		}
	}

	return false
}
