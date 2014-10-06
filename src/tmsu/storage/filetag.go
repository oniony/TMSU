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
func (storage *Storage) FileTagExists(fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId, explicitOnly bool) (bool, error) {
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
func (storage *Storage) FileTagCountByFileId(fileId entities.FileId, explicitOnly bool) (uint, error) {
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
func (storage *Storage) FileTagCountByTagId(tagId entities.TagId, explicitOnly bool) (uint, error) {
	if explicitOnly {
		return storage.Db.FileTagCountByTagId(tagId)
	}

	fileTags, err := storage.FileTagsByTagId(tagId, false)
	if err != nil {
		return 0, err
	}

	return uint(len(fileTags)), err
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId entities.TagId, explicitOnly bool) (entities.FileTags, error) {
	fileTags, err := storage.Db.FileTagsByTagId(tagId)
	if err != nil {
		return nil, err
	}

	if !explicitOnly {
		var err error
		fileTags, err = storage.addImpliedFileTags(fileTags)
		if err != nil {
			return nil, err
		}
	}

	return fileTags, nil
}

// Retrieves the count of file tags for the specified value.
func (storage *Storage) FileTagCountByValueId(valueId entities.ValueId) (uint, error) {
	return storage.Db.FileTagCountByValueId(valueId)
}

// Retrieves the file tags with the specified value ID.
func (storage *Storage) FileTagsByValueId(valueId entities.ValueId) (entities.FileTags, error) {
	return storage.Db.FileTagsByValueId(valueId)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId entities.FileId, explicitOnly bool) (entities.FileTags, error) {
	fileTags, err := storage.Db.FileTagsByFileId(fileId)
	if err != nil {
		return nil, err
	}

	if !explicitOnly {
		var err error
		fileTags, err = storage.addImpliedFileTags(fileTags)
		if err != nil {
			return nil, err
		}
	}

	return fileTags, nil
}

// Adds a file tag.
func (storage *Storage) AddFileTag(fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId) (*entities.FileTag, error) {
	return storage.Db.AddFileTag(fileId, tagId, valueId)
}

// Delete file tag.
func (storage *Storage) DeleteFileTag(fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId) error {
	exists, err := storage.FileTagExists(fileId, tagId, valueId, true)
	if err != nil {
		return err
	}
	if !exists {
		return FileTagDoesNotExist{fileId, tagId, valueId}
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
func (storage *Storage) DeleteFileTagsByFileId(fileId entities.FileId) error {
	fileTags, err := storage.Db.FileTagsByFileId(fileId)
	if err != nil {
		return err
	}

	if err := storage.Db.DeleteFileTagsByFileId(fileId); err != nil {
		return err
	}

	if err := storage.DeleteFileIfUntagged(fileId); err != nil {
		return err
	}

	if err := storage.DeleteUnusedValues(fileTags.ValueIds()); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified tag.
func (storage *Storage) DeleteFileTagsByTagId(tagId entities.TagId) error {
	fileTags, err := storage.Db.FileTagsByTagId(tagId)
	if err != nil {
		return err
	}

	if err := storage.Db.DeleteFileTagsByTagId(tagId); err != nil {
		return err
	}

	if err := storage.DeleteUntaggedFiles(fileTags.FileIds()); err != nil {
		return err
	}

	if err := storage.DeleteUnusedValues(fileTags.ValueIds()); err != nil {
		return err
	}

	return nil
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId, destTagId entities.TagId) error {
	return storage.Db.CopyFileTags(sourceTagId, destTagId)
}

// unexported

func (storage *Storage) addImpliedFileTags(fileTags entities.FileTags) (entities.FileTags, error) {
	tagIds := make(entities.TagIds, 0, len(fileTags))
	for _, fileTag := range fileTags {
		tagIds = append(tagIds, fileTag.TagId)
	}

	implications, err := storage.ImplicationsForTags(tagIds...)
	if err != nil {
		return nil, err
	}

	for index := 0; index < len(fileTags); index++ {
		fileTag := fileTags[index]

		for _, implication := range implications {
			if implication.ImplyingTag.Id == fileTag.TagId {
				//TODO consider values in implied tags
				impliedFileTag := fileTags.Find(fileTag.FileId, implication.ImpliedTag.Id, 0)
				if impliedFileTag != nil {
					impliedFileTag.Implicit = true
				} else {
					impliedFileTag := entities.FileTag{fileTag.FileId, implication.ImpliedTag.Id, 0, false, true}
					fileTags = append(fileTags, &impliedFileTag)
				}
			}
		}
	}

	return fileTags, nil
}
