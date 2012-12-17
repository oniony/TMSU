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
	"errors"
	"path/filepath"
	"sort"
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
	explicitlyTaggedFiles, err := storage.Db.FilesWithTag(tagId)
	if err != nil {
		return nil, err
	}

	if explicitOnly {
		return explicitlyTaggedFiles, nil
	}

	files := make(database.Files, len(explicitlyTaggedFiles))
	for index, file := range explicitlyTaggedFiles {
		files[index] = file
	}

	for _, explicitlyTaggedFile := range explicitlyTaggedFiles {
		additionalFiles, err := storage.Db.FilesByDirectory(explicitlyTaggedFile.Path())
		if err != nil {
			return nil, err
		}

		for _, additionalFile := range additionalFiles {
			files = append(files, additionalFile)
		}
	}

	return files, nil
}

// Retrieves the set of files with the specified set of tags.
func (storage *Storage) FilesWithTags(includeTagIds, excludeTagIds []uint, explicitOnly bool) (database.Files, error) {
	var files database.Files
	var err error

	if len(includeTagIds) > 0 {
		files, err = storage.FilesWithTag(includeTagIds[0], explicitOnly)
		if err != nil {
			return nil, err
		}

		for _, tagId := range includeTagIds[1:] {
			filesWithTag, err := storage.FilesWithTag(tagId, explicitOnly)
			if err != nil {
				return nil, err
			}

			for index, file := range files {
				if file == nil {
					continue
				}

				if !contains(filesWithTag, file) {
					files[index] = nil
				}
			}
		}
	}

	if len(excludeTagIds) > 0 {
		if len(includeTagIds) == 0 {
			files, err = storage.Db.Files()
			if err != nil {
				return nil, err
			}
		}

		for _, tagId := range excludeTagIds {
			filesWithTag, err := storage.FilesWithTag(tagId, explicitOnly)
			if err != nil {
				return nil, err
			}

			for index, file := range files {
				if file == nil {
					continue
				}

				if contains(filesWithTag, file) {
					files[index] = nil
				}
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
func (storage *Storage) FileTagCount() (uint, error) {
	return storage.Db.FileTagCount()
}

// Retrieves the complete set of file tags.
func (storage *Storage) FileTags() (database.FileTags, error) {
	return storage.Db.FileTags()
}

// Retrieves the set of explicit file tags for the specified file.
func (storage *Storage) TagsByFileId(fileId uint, explicitOnly bool) (database.Tags, error) {
	tags, err := storage.Db.TagsByFileId(fileId)
	if err != nil {
		return nil, err
	}

	if explicitOnly {
		return tags, nil
	}

	file, err := storage.File(fileId)
	if err != nil {
		return nil, err
	}

	absPath := filepath.Dir(file.Path())
	for absPath != "/" {
		file, err = storage.Db.FileByPath(absPath)
		if err != nil {
			return nil, err
		}

		if file != nil {
			moreTags, err := storage.Db.TagsByFileId(file.Id)
			if err != nil {
				return nil, err
			}
			tags = append(tags, moreTags...)
		}

		absPath = filepath.Dir(absPath)
	}

	sort.Sort(tags)
	tags = uniq(tags)

	return tags, nil
}

// Retrieves the set of tags for the specified path.
func (storage *Storage) TagsForPath(path string, explicitOnly bool) (database.Tags, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := storage.Db.FileByPath(absPath)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return database.Tags{}, nil
	}

	return storage.TagsByFileId(file.Id, explicitOnly)
}

// Retrieves the file tag with the specified file ID and tag ID.
func (storage *Storage) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*database.FileTag, error) {
	return storage.Db.FileTagByFileIdAndTagId(fileId, tagId)
}

// Retrieves the file tags with the specified tag ID.
func (storage *Storage) FileTagsByTagId(tagId uint, explicitOnly bool) (database.FileTags, error) {
	if !explicitOnly {
		panic("Not implemented.")
	} //TODO implement implicit

	return storage.Db.FileTagsByTagId(tagId)
}

// Retrieves the file tags with the specified file ID.
func (storage *Storage) FileTagsByFileId(fileId uint, explicitOnly bool) (database.FileTags, error) {
	fileTags, err := storage.Db.FileTagsByFileId(fileId)
	if err != nil {
		return nil, err
	}

	if explicitOnly {
		return fileTags, nil
	}

	file, err := storage.Db.File(fileId)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("File does not exist.")
	}

	parentFile, err := storage.Db.FileByPath(file.Directory)
	if err != nil {
		return nil, err
	}

	if parentFile != nil {
		additionalFileTags, err := storage.FileTagsByFileId(parentFile.Id, explicitOnly)
		if err != nil {
			return nil, err
		}

		for _, additionalFileTag := range additionalFileTags {
			fileTags = append(fileTags, additionalFileTag)
		}
	}

	return fileTags, nil
}

// Adds a file tag.
func (storage *Storage) AddFileTag(fileId uint, tagId uint) (*database.FileTag, error) {
	return storage.Db.AddFileTag(fileId, tagId)
}

// Removes a file tag.
func (storage *Storage) RemoveFileTag(fileId uint, tagId uint) error {
	return storage.Db.RemoveFileTag(fileId, tagId)
}

// Removes all of the file tags for the specified file.
func (storage *Storage) RemoveFileTagsByFileId(fileId uint) error {
	return storage.Db.RemoveFileTagsByFileId(fileId)
}

// Removes all of the file tags for the specified tag.
func (storage *Storage) RemoveFileTagsByTagId(tagId uint) error {
	return storage.Db.RemoveFileTagsByTagId(tagId)
}

// Updates file tags to a new tag.
func (storage *Storage) UpdateFileTags(oldTagId uint, newTagId uint) error {
	return storage.Db.UpdateFileTags(oldTagId, newTagId)
}

// Copies file tags from one tag to another.
func (storage *Storage) CopyFileTags(sourceTagId uint, destTagId uint) error {
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

func uniq(tags database.Tags) database.Tags {
	uniqueTags := make(database.Tags, 0, len(tags))

	var previousTagName string = ""
	for _, tag := range tags {
		if tag.Name == previousTagName {
			continue
		}

		uniqueTags = append(uniqueTags, tag)
		previousTagName = tag.Name
	}

	return uniqueTags
}
