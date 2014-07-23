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

// Determines whether the specified file has the specified tag applied.
func (storage *Storage) FileTagExists(fileId, tagId, valueId uint, explicitOnly bool) (bool, error) {
	if explicitOnly {
		return storage.Db.FileTagExists(fileId, tagId, valueId)
	}

	fileTags, err := storage.FileTagsByFileId(fileId, false)
	if err != nil {
		return false, err
	}

	return fileTags.Contains(tagId, valueId), nil
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
func (storage *Storage) FileTagCountByFileId(fileId uint, explicitOnly bool) (uint, error) {
	if explicitOnly {
		return storage.Db.FileTagCountByFileId(fileId)
	}

	fileTags, err := storage.FileTagsByFileId(fileId, false)
	if err != nil {
		return 0, err
	}

	return uint(len(fileTags)), err
}

// Retrieves the count of file tags for the specified tag.
func (storage *Storage) FileTagCountByTagId(tagId uint) (uint, error) {
	//TODO add explicit only
	return storage.Db.FileTagCountByTagId(tagId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint) (entities.FileTags, error) {
	//TODO add explicit only
	return storage.Db.FileTagsByTagId(tagId)
}

// Retrieves the count of file tags for the specified value.
func (storage *Storage) FileTagCountByValueId(valueId uint) (uint, error) {
	return storage.Db.FileTagCountByValueId(valueId)
}

// Retrieves the file tags with the specified value ID.
func (storage *Storage) FileTagsByValueId(valueId uint) (entities.FileTags, error) {
	return storage.Db.FileTagsByValueId(valueId)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId uint, explicitOnly bool) (entities.FileTags, error) {
	fileTags, err := storage.Db.FileTagsByFileId(fileId)
	if err != nil {
		return nil, err
	}

	if !explicitOnly {
		tagIds := make([]uint, 0, len(fileTags))
		for _, fileTag := range fileTags {
			tagIds = append(tagIds, fileTag.TagId)
		}

		implications, err := storage.ImplicationsForTags(tagIds...)
		if err != nil {
			return nil, err
		}

		for _, implication := range implications {
			fileTag := findFileTag(fileTags, implication.ImpliedTag.Id)
			if fileTag != nil {
				fileTag.Implicit = true
			} else {
				impliedFileTag := entities.FileTag{fileId, implication.ImpliedTag.Id, 0, false, true}
				fileTags = append(fileTags, &impliedFileTag)
			}
		}
	}

	return fileTags, nil
}

// Adds a file tag.
func (storage *Storage) AddFileTag(fileId, tagId, valueId uint) (*entities.FileTag, error) {
	return storage.Db.AddFileTag(fileId, tagId, valueId)
}

// Delete file tag.
func (storage *Storage) DeleteFileTag(fileId, tagId, valueId uint) error {
	exists, err := storage.FileTagExists(fileId, tagId, valueId, true)
	if err != nil {
		return err
	}
	if !exists {
		return FileTagDoesNotExist
	}

	if err := storage.Db.DeleteFileTag(fileId, tagId, valueId); err != nil {
		return err
	}

	if err := storage.DeleteFileIfUntagged(fileId); err != nil {
		return err
	}

	if err := storage.DeleteValueIfUnused(valueId); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified file.
func (storage *Storage) DeleteFileTagsByFileId(fileId uint) error {
	if err := storage.Db.DeleteFileTagsByFileId(fileId); err != nil {
		return err
	}

	if err := storage.DeleteFileIfUntagged(fileId); err != nil {
		return err
	}

	//TODO look only at the values that were in the filetags removed
	if err := storage.DeleteUnusedValues(); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified tag.
func (storage *Storage) DeleteFileTagsByTagId(tagId uint) error {
	if err := storage.Db.DeleteFileTagsByTagId(tagId); err != nil {
		return err
	}

	if err := storage.DeleteUntaggedFiles(); err != nil {
		return err
	}

	if err := storage.DeleteUnusedValues(); err != nil {
		return err
	}

	return nil
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId, destTagId uint) error {
	return storage.Db.CopyFileTags(sourceTagId, destTagId)
}

// unexported

func findFileTag(fileTags entities.FileTags, tagId uint) *entities.FileTag {
	for _, fileTag := range fileTags {
		if fileTag.TagId == tagId {
			return fileTag
		}
	}

	return nil
}
