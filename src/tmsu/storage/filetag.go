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
	"tmsu/entities"
)

// Determines whether the specified file has the specified tag applied.
func (storage *Storage) FileTagExists(fileId, tagId, valueId uint) (bool, error) {
	return storage.Db.FileTagExists(fileId, tagId, valueId)
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

// Retrieves the count of file tags for the specified tag.
func (storage *Storage) FileTagCountByTagId(tagId uint) (uint, error) {
	return storage.Db.FileTagCountByTagId(tagId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint) (entities.FileTags, error) {
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
func (storage *Storage) FileTagsByFileId(fileId uint) (entities.FileTags, error) {
	return storage.Db.FileTagsByFileId(fileId)
}

// Adds a file tag.
func (storage *Storage) AddFileTag(fileId, tagId, valueId uint) (*entities.FileTag, error) {
	return storage.Db.AddFileTag(fileId, tagId, valueId)
}

// Delete file tag.
func (storage *Storage) DeleteFileTag(fileId, tagId, valueId uint) error {
	if err := storage.Db.DeleteFileTag(fileId, tagId, valueId); err != nil {
		return err
	}

	count, err := storage.Db.FileTagCountByFileId(fileId)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.Db.DeleteFile(fileId); err != nil {
			return err
		}
	}

	count, err = storage.Db.FileTagCountByValueId(valueId)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.Db.DeleteValue(valueId); err != nil {
			return err
		}
	}

	return nil
}

// Deletes all of the file tags for the specified file.
func (storage *Storage) DeleteFileTagsByFileId(fileId uint) error {
	fileTags, err := storage.Db.FileTagsByFileId(fileId)
	if err != nil {
		return err
	}

	if err := storage.Db.DeleteFileTagsByFileId(fileId); err != nil {
		return err
	}

	count, err := storage.Db.FileTagCountByFileId(fileId)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.Db.DeleteFile(fileId); err != nil {
			return err
		}
	}

	for _, fileTag := range fileTags {
		count, err := storage.Db.FileTagCountByValueId(fileTag.ValueId)
		if err != nil {
			return err
		}
		if count == 0 {
			if err := storage.Db.DeleteValue(fileTag.ValueId); err != nil {
				return err
			}
		}
	}

	return nil
}

// Deletes all of the file tags for the specified tag.
func (storage *Storage) DeleteFileTagsByTagId(tagId uint) error {
	fileTags, err := storage.Db.FileTagsByTagId(tagId)
	if err != nil {
		return err
	}

	if err := storage.Db.DeleteFileTagsByTagId(tagId); err != nil {
		return err
	}

	for _, fileTag := range fileTags {
		count, err := storage.Db.FileTagCountByFileId(fileTag.FileId)
		if err != nil {
			return err
		}
		if count == 0 {
			if err := storage.Db.DeleteFile(fileTag.FileId); err != nil {
				return err
			}
		}

		count, err = storage.Db.FileTagCountByValueId(fileTag.ValueId)
		if err != nil {
			return err
		}
		if count == 0 {
			if err := storage.Db.DeleteValue(fileTag.ValueId); err != nil {
				return err
			}
		}
	}

	return nil
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
