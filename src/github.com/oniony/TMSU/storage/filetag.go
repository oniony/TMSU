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

// Determines whether the specified file has the specified tag applied.
func (storage *Storage) FileTagExists(tx *Tx, fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId, explicitOnly bool) (bool, error) {
	if explicitOnly {
		return database.FileTagExists(tx.tx, fileId, tagId, valueId)
	}

	fileTags, err := storage.FileTagsByFileId(tx, fileId, false)
	if err != nil {
		return false, err
	}

	predicate := func(fileTag entities.FileTag) bool {
		return fileTag.TagId == tagId && fileTag.ValueId == valueId
	}

	return fileTags.Any(predicate), nil
}

// Retrieves the total count of file tags in the database.
func (storage *Storage) FileTagCount(tx *Tx) (uint, error) {
	return database.FileTagCount(tx.tx)
}

// Retrieves the complete set of file tags.
func (storage *Storage) FileTags(tx *Tx) (entities.FileTags, error) {
	return database.FileTags(tx.tx)
}

// Retrieves the count of file tags for the specified file.
func (storage *Storage) FileTagCountByFileId(tx *Tx, fileId entities.FileId, explicitOnly bool) (uint, error) {
	if explicitOnly {
		return database.FileTagCountByFileId(tx.tx, fileId)
	}

	fileTags, err := storage.FileTagsByFileId(tx, fileId, false)
	if err != nil {
		return 0, err
	}

	return uint(len(fileTags)), err
}

// Retrieves the count of file tags for the specified tag.
func (storage *Storage) FileTagCountByTagId(tx *Tx, tagId entities.TagId, explicitOnly bool) (uint, error) {
	if explicitOnly {
		return database.FileTagCountByTagId(tx.tx, tagId)
	}

	fileTags, err := storage.FileTagsByTagId(tx, tagId, false)
	if err != nil {
		return 0, err
	}

	return uint(len(fileTags)), err
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tx *Tx, tagId entities.TagId, explicitOnly bool) (entities.FileTags, error) {
	fileTags, err := database.FileTagsByTagId(tx.tx, tagId)
	if err != nil {
		return nil, err
	}

	if !explicitOnly {
		var err error
		fileTags, err = storage.addImpliedFileTags(tx, fileTags)
		if err != nil {
			return nil, err
		}
	}

	return fileTags, nil
}

// Retrieves the count of file tags for the specified value.
func (storage *Storage) FileTagCountByValueId(tx *Tx, valueId entities.ValueId) (uint, error) {
	return database.FileTagCountByValueId(tx.tx, valueId)
}

// Retrieves the file tags with the specified value ID.
func (storage *Storage) FileTagsByValueId(tx *Tx, valueId entities.ValueId) (entities.FileTags, error) {
	return database.FileTagsByValueId(tx.tx, valueId)
}

// Retrieves the file tags for the specified file ID.
func (storage *Storage) FileTagsByFileId(tx *Tx, fileId entities.FileId, explicitOnly bool) (entities.FileTags, error) {
	fileTags, err := database.FileTagsByFileId(tx.tx, fileId)
	if err != nil {
		return nil, err
	}

	if !explicitOnly {
		var err error
		fileTags, err = storage.addImpliedFileTags(tx, fileTags)
		if err != nil {
			return nil, err
		}
	}

	return fileTags, nil
}

// Adds a file tag.
func (storage *Storage) AddFileTag(tx *Tx, fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId) (*entities.FileTag, error) {
	return database.AddFileTag(tx.tx, fileId, tagId, valueId)
}

// Delete file tag.
func (storage *Storage) DeleteFileTag(tx *Tx, fileId entities.FileId, tagId entities.TagId, valueId entities.ValueId) error {
	exists, err := storage.FileTagExists(tx, fileId, tagId, valueId, true)
	if err != nil {
		return err
	}
	if !exists {
		return FileTagDoesNotExist{fileId, tagId, valueId}
	}

	if err := database.DeleteFileTag(tx.tx, fileId, tagId, valueId); err != nil {
		return err
	}

	if err := storage.DeleteFileIfUntagged(tx, fileId); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified file.
func (storage *Storage) DeleteFileTagsByFileId(tx *Tx, fileId entities.FileId) error {
	if err := database.DeleteFileTagsByFileId(tx.tx, fileId); err != nil {
		return err
	}

	if err := storage.DeleteFileIfUntagged(tx, fileId); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified tag.
func (storage *Storage) DeleteFileTagsByTagId(tx *Tx, tagId entities.TagId) error {
	fileTags, err := database.FileTagsByTagId(tx.tx, tagId)
	if err != nil {
		return err
	}

	if err := database.DeleteFileTagsByTagId(tx.tx, tagId); err != nil {
		return err
	}

	if err := storage.DeleteUntaggedFiles(tx, fileTags.FileIds()); err != nil {
		return err
	}

	return nil
}

// Deletes all of the file tags for the specified value.
func (storage *Storage) DeleteFileTagsByValueId(tx *Tx, valueId entities.ValueId) error {
	fileTags, err := database.FileTagsByValueId(tx.tx, valueId)
	if err != nil {
		return err
	}

	if err := database.DeleteFileTagsByValueId(tx.tx, valueId); err != nil {
		return err
	}

	if err := storage.DeleteUntaggedFiles(tx, fileTags.FileIds()); err != nil {
		return err
	}

	return nil
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(tx *Tx, sourceTagId, destTagId entities.TagId) error {
	return database.CopyFileTags(tx.tx, sourceTagId, destTagId)
}

// unexported

func (storage *Storage) addImpliedFileTags(tx *Tx, fileTags entities.FileTags) (entities.FileTags, error) {
	// WARN: this cannot use 'range' as fileTags is expanded within the loop
	for index := 0; index < len(fileTags); index++ {
		fileTag := fileTags[index]

		implications, err := storage.ImplicationsFor(tx, fileTag.ToTagIdValueIdPair())
		if err != nil {
			return nil, err
		}

		for _, implication := range implications {
			predicate := func(ft entities.FileTag) bool {
				return ft.FileId == fileTag.FileId &&
					ft.TagId == implication.ImpliedTag.Id &&
					ft.ValueId == implication.ImpliedValue.Id
			}

			impliedFileTag := fileTags.Where(predicate).Single()

			if impliedFileTag != nil {
				impliedFileTag.Implicit = true
			} else {
				impliedFileTag := entities.FileTag{fileTag.FileId, implication.ImpliedTag.Id, implication.ImpliedValue.Id, false, true}

				fileTags = append(fileTags, &impliedFileTag)
			}
		}
	}

	return fileTags, nil
}
