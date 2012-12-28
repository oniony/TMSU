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
	"fmt"
	"tmsu/storage/database"
)

func (storage *Storage) FileCountWithTags(tagIds []uint, explicitOnly bool) (uint, error) {
	//TODO optimize
	files, err := storage.FilesWithTags(tagIds, []uint{}, explicitOnly)
	if err != nil {
		return 0, err
	}

	return uint(len(files)), nil
}

// Retrieves the set of files with the specified tag.
func (storage *Storage) FilesWithTag(tagId uint, explicitOnly bool) (database.Files, error) {
	return storage.Db.FilesWithTag(tagId, explicitOnly)
}

// Retrieves the set of files with the specified set of tags.
func (storage *Storage) FilesWithTags(includeTagIds, excludeTagIds []uint, explicitOnly bool) (database.Files, error) {
	var files database.Files
	var err error

	if len(includeTagIds) > 0 {
		files, err = storage.Db.FilesWithTags(includeTagIds, explicitOnly)
		if err != nil {
			return nil, err
		}
	}

	if len(excludeTagIds) > 0 {
		if len(includeTagIds) == 0 {
			files, err = storage.Db.Files()
			if err != nil {
				return nil, err
			}
		}

		excludeFiles, err := storage.Db.FilesWithTags(excludeTagIds, explicitOnly)
		if err != nil {
			return nil, err
		}

		for index, file := range files {
			if contains(excludeFiles, file) {
				files[index] = nil
			}
		}
	}

	resultFiles := make(database.Files, 0, len(files))
	for _, file := range files {
		if file != nil {
			resultFiles = append(resultFiles, file)
		}
	}

	return resultFiles, nil
}

// Retrieves the total count of file tags in the database.
func (storage *Storage) FileTagCount(explicitOnly bool) (uint, error) {
	return storage.Db.FileTagCount(explicitOnly)
}

// Retrieves the complete set of file tags.
func (storage *Storage) FileTags(explicitOnly bool) (database.FileTags, error) {
	return storage.Db.FileTags(explicitOnly)
}

// Retrieves the count of file tags for the specified file.
func (storage *Storage) FileTagCountByFileId(fileId uint, explicitOnly bool) (uint, error) {
	return storage.Db.FileTagCountByFileId(fileId, explicitOnly)
}

// Retrieves the set of file tags for the specified file.
func (storage *Storage) TagsByFileId(fileId uint, explicitOnly bool) (database.Tags, error) {
	return storage.Db.TagsByFileId(fileId, explicitOnly)
}

// Retrieves the file tag with the specified file ID and tag ID.
func (storage *Storage) FileTagByFileIdAndTagId(fileId, tagId uint) (*database.FileTag, error) {
	return storage.Db.FileTagByFileIdAndTagId(fileId, tagId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint, explicitOnly bool) (database.FileTags, error) {
	return storage.Db.FileTagsByTagId(tagId, explicitOnly)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId uint, explicitOnly bool) (database.FileTags, error) {
	return storage.Db.FileTagsByFileId(fileId, explicitOnly)
}

// Adds an explicit file tag.
func (storage *Storage) AddExplicitFileTag(fileId, tagId uint) (*database.FileTag, error) {
	fileTag, err := storage.Db.FileTagByFileIdAndTagId(fileId, tagId)
	if err != nil {
		return nil, err
	}

	if fileTag == nil {
		fileTag, err = storage.Db.InsertFileTag(fileId, tagId, true, false)
	} else {
		fileTag, err = storage.Db.UpdateFileTag(fileTag.Id, fileTag.FileId, fileTag.TagId, true, fileTag.Implicit)
	}

	if err != nil {
		return nil, err
	}

	return fileTag, nil
}

// Adds an impicit file tag.
func (storage *Storage) AddImplicitFileTag(fileId, tagId uint) (*database.FileTag, error) {
	fileTag, err := storage.Db.FileTagByFileIdAndTagId(fileId, tagId)
	if err != nil {
		return nil, err
	}

	if fileTag == nil {
		fileTag, err = storage.Db.InsertFileTag(fileId, tagId, false, true)
	} else {
		fileTag, err = storage.Db.UpdateFileTag(fileTag.Id, fileTag.FileId, fileTag.TagId, fileTag.Explicit, true)
	}

	if err != nil {
		return nil, err
	}

	return fileTag, nil
}

// Adds a file tag.
func (storage *Storage) AddFileTag(fileId, tagId uint, explicit, implicit bool) (*database.FileTag, error) {
	fileTag, err := storage.Db.FileTagByFileIdAndTagId(fileId, tagId)
	if err != nil {
		return nil, err
	}
	//TODO may be nil

	if fileTag == nil {
		fileTag, err = storage.Db.InsertFileTag(fileId, tagId, explicit, implicit)
		if err != nil {
			return nil, err
		}
	} else {
		newExplicit := fileTag.Explicit || explicit
		newImplicit := fileTag.Implicit || implicit

		fileTag, err = storage.Db.UpdateFileTag(fileTag.Id, fileTag.FileId, fileTag.TagId, newExplicit, newImplicit)
		if err != nil {
			return nil, err
		}
	}

	return fileTag, err
}

// Remove explicit file tag.
func (storage *Storage) RemoveExplicitFileTag(fileTagId uint) error {
	fileTag, err := storage.Db.FileTag(fileTagId)
	if err != nil {
		return err
	}
	if fileTag == nil {
		return errors.New(fmt.Sprintf("No such file tag '%v'."))
	}

	if fileTag.Implicit {
		_, err = storage.Db.UpdateFileTag(fileTagId, fileTag.FileId, fileTag.TagId, false, true)
	} else {
		err = storage.Db.DeleteFileTag(fileTagId)
	}

	if err != nil {
		return err
	}

	return nil
}

// Remove implicit file tag.
func (storage *Storage) RemoveImplicitFileTag(fileTagId uint) error {
	fileTag, err := storage.Db.FileTag(fileTagId)
	if err != nil {
		return err
	}
	if fileTag == nil {
		return errors.New(fmt.Sprintf("No such file tag '%v'."))
	}

	if fileTag.Explicit {
		_, err = storage.Db.UpdateFileTag(fileTagId, fileTag.FileId, fileTag.TagId, true, false)
	} else {
		err = storage.Db.DeleteFileTag(fileTagId)
	}

	if err != nil {
		return err
	}

	return nil
}

// Removes a file tag.
func (storage *Storage) RemoveFileTag(fileTagId uint) error {
	return storage.Db.DeleteFileTag(fileTagId)
}

// Removes a file tag by file and tag ID.
func (storage *Storage) RemoveFileTagByFileAndTagId(fileId, tagId uint) error {
	return storage.Db.DeleteFileTagByFileAndTagId(fileId, tagId)
}

// Removes all of the file tags for the specified file.
func (storage *Storage) RemoveFileTagsByFileId(fileId uint, explicitOnly bool) error {
	return storage.Db.DeleteFileTagsByFileId(fileId, explicitOnly)
}

// Removes all of the file tags for the specified tag.
func (storage *Storage) RemoveFileTagsByTagId(tagId uint, explicitOnly bool) error {
	return storage.Db.DeleteFileTagsByTagId(tagId, explicitOnly)
}

// Updates a file tag.
func (storage *Storage) UpdateFileTag(fileTagId, fileId, tagId uint, explicit, implicit bool) (*database.FileTag, error) {
	return storage.Db.UpdateFileTag(fileTagId, fileId, tagId, explicit, implicit)
}

// Updates file tags to a new tag.
func (storage *Storage) UpdateFileTags(oldTagId, newTagId uint) error {
	return storage.Db.UpdateFileTags(oldTagId, newTagId)
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId, destTagId uint) error {
	return storage.Db.CopyFileTags(sourceTagId, destTagId)
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
