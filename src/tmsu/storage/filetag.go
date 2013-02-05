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
	"tmsu/storage/database"
)

// Determines whether the specified file has the specified tag applied.
func (storage *Storage) FileTagExists(fileId, tagId uint) (bool, error) {
	return storage.Db.FileTagExists(fileId, tagId)
}

// Determines whether the specified file has the specified explicit tag applied.
func (storage *Storage) ExplicitFileTagExists(fileId, tagId uint) (bool, error) {
	return storage.Db.ExplicitFileTagExists(fileId, tagId)
}

// Determines whether the specified file has the specified implicit tag applied.
func (storage *Storage) ImplicitFileTagExists(fileId, tagId uint) (bool, error) {
	return storage.Db.ImplicitFileTagExists(fileId, tagId)
}

// Retrieves the total count of file tags in the database.
func (storage *Storage) FileTagCount() (uint, error) {
	return storage.Db.FileTagCount()
}

// Retrieves the total count of explicit file tags in the database.
func (storage *Storage) ExplicitFileTagCount() (uint, error) {
	return storage.Db.ExplicitFileTagCount()
}

// Retrieves the total count of implicit file tags in the database.
func (storage *Storage) ImplicitFileTagCount() (uint, error) {
	return storage.Db.ImplicitFileTagCount()
}

// Retrieves the complete set of file tags.
func (storage *Storage) FileTags() (database.FileTags, error) {
	return storage.Db.FileTags()
}

// Retrieves the complete set of explicit file tags.
func (storage *Storage) ExplicitFileTags() (database.FileTags, error) {
	return storage.Db.ExplicitFileTags()
}

// Retrieves the complete set of implicit file tags.
func (storage *Storage) ImplicitFileTags() (database.FileTags, error) {
	return storage.Db.ImplicitFileTags()
}

// Retrieves the count of file tags for the specified file.
func (storage *Storage) FileTagCountByFileId(fileId uint) (uint, error) {
	return storage.Db.FileTagCountByFileId(fileId)
}

// Retrieves the count of explicit file tags for the specified file.
func (storage *Storage) ExplicitFileTagCountByFileId(fileId uint) (uint, error) {
	return storage.Db.ExplicitFileTagCountByFileId(fileId)
}

// Retrieves the count of implicit file tags for the specified file.
func (storage *Storage) ImplicitFileTagCountByFileId(fileId uint) (uint, error) {
	return storage.Db.ImplicitFileTagCountByFileId(fileId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint) (database.FileTags, error) {
	return storage.Db.FileTagsByTagId(tagId)
}

// Retrieves the explicit file tags with the specified tag ID.
func (storage *Storage) ExplicitFileTagsByTagId(tagId uint) (database.FileTags, error) {
	return storage.Db.ExplicitFileTagsByTagId(tagId)
}

// Retrieves the implicit file tags with the specified tag ID.
func (storage *Storage) ImplicitFileTagsByTagId(tagId uint) (database.FileTags, error) {
	return storage.Db.ImplicitFileTagsByTagId(tagId)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId uint) (database.FileTags, error) {
	return storage.Db.FileTagsByFileId(fileId)
}

// Retrieves the explicit file tags with the specified file ID.
func (storage *Storage) ExplicitFileTagsByFileId(fileId uint) (database.FileTags, error) {
	return storage.Db.ExplicitFileTagsByFileId(fileId)
}

// Retrieves the implicit file tags with the specified file ID.
func (storage *Storage) ImplicitFileTagsByFileId(fileId uint) (database.FileTags, error) {
	return storage.Db.ImplicitFileTagsByFileId(fileId)
}

// Adds an explicit file tag.
func (storage *Storage) AddExplicitFileTag(fileId, tagId uint) (*database.FileTag, error) {
	return storage.Db.AddExplicitFileTag(fileId, tagId)
}

// Adds an impicit file tag.
func (storage *Storage) AddImplicitFileTag(fileId, tagId uint) (*database.FileTag, error) {
	return storage.Db.AddImplicitFileTag(fileId, tagId)
}

// Adds a set of explicit file tags.
func (storage *Storage) AddExplicitFileTags(fileId uint, tagIds []uint) error {
	if len(tagIds) == 0 {
		return nil
	}

	return storage.Db.AddExplicitFileTags(fileId, tagIds)
}

// Adds a set of implicit file tags.
func (storage *Storage) AddImplicitFileTags(fileId uint, tagIds []uint) error {
	if len(tagIds) == 0 {
		return nil
	}

	return storage.Db.AddImplicitFileTags(fileId, tagIds)
}

// Remove explicit file tag.
func (storage *Storage) RemoveExplicitFileTag(fileId, tagId uint) error {
	return storage.Db.DeleteExplicitFileTag(fileId, tagId)
}

// Remove implicit file tag.
func (storage *Storage) RemoveImplicitFileTag(fileId, tagId uint) error {
	return storage.Db.DeleteImplicitFileTag(fileId, tagId)
}

// Removes all of the file tags for the specified file.
func (storage *Storage) RemoveFileTagsByFileId(fileId uint) error {
	err := storage.Db.DeleteImplicitFileTagsByFileId(fileId)
	if err != nil {
		return err
	}

	return storage.Db.DeleteExplicitFileTagsByFileId(fileId)
}

// Removes all of the explicit file tags for the specified file.
func (storage *Storage) RemoveExplicitFileTagsByFileId(fileId uint) error {
	return storage.Db.DeleteExplicitFileTagsByFileId(fileId)
}

// Removes all of the implicit file tags for the specified file.
func (storage *Storage) RemoveImplicitFileTagsByFileId(fileId uint) error {
	return storage.Db.DeleteImplicitFileTagsByFileId(fileId)
}

// Removes all of the file tags for the specified tag.
func (storage *Storage) RemoveFileTagsByTagId(tagId uint) error {
	err := storage.Db.DeleteImplicitFileTagsByTagId(tagId)
	if err != nil {
		return err
	}

	return storage.Db.DeleteExplicitFileTagsByTagId(tagId)
}

// Removes all of the explicit file tags for the specified tag.
func (storage *Storage) RemoveExplicitFileTagsByTagId(tagId uint) error {
	return storage.Db.DeleteExplicitFileTagsByTagId(tagId)
}

// Removes all of the implicit file tags for the specified tag.
func (storage *Storage) RemoveImplicitFileTagsByTagId(tagId uint) error {
	return storage.Db.DeleteImplicitFileTagsByTagId(tagId)
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId, destTagId uint) error {
	err := storage.Db.CopyExplicitFileTags(sourceTagId, destTagId)
	if err != nil {
		return err
	}

	return storage.Db.CopyImplicitFileTags(sourceTagId, destTagId)
}

// helpers

func contains(files database.Files, searchFile *database.File) bool {
	for _, file := range files {
		if file.Path() == searchFile.Path() {
			return true
		}
	}

	return false
}
