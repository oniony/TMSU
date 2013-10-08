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
	"fmt"
	"time"
	"tmsu/fingerprint"
	"tmsu/query"
	"tmsu/storage/database"
)

// Retrieves the total number of tracked files.
func (storage *Storage) FileCount() (uint, error) {
	return storage.Db.FileCount()
}

// The complete set of tracked files.
func (storage *Storage) Files() (database.Files, error) {
	return storage.Db.Files()
}

// Retrieves a specific file.
func (storage *Storage) File(id uint) (*database.File, error) {
	return storage.Db.File(id)
}

// Retrieves the file with the specified path.
func (storage *Storage) FileByPath(path string) (*database.File, error) {
	return storage.Db.FileByPath(path)
}

// Retrieves all files that are under the specified directory.
func (storage *Storage) FilesByDirectory(path string) (database.Files, error) {
	return storage.Db.FilesByDirectory(path)
}

// Retrieves all file that are under the specified directories.
func (storage *Storage) FilesByDirectories(paths []string) (database.Files, error) {
	files := make(database.Files, 0, 100)

	for _, path := range paths {
		pathFiles, err := storage.Db.FilesByDirectory(path)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve files for directory: %v", path, err)
		}

		files = append(files, pathFiles...)
	}

	return files, nil
}

// Retrieves the number of files with the specified fingerprint.
func (storage *Storage) FileCountByFingerprint(fingerprint fingerprint.Fingerprint) (uint, error) {
	return storage.Db.FileCountByFingerprint(fingerprint)
}

// Retrieves the set of files with the specified fingerprint.
func (storage *Storage) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (database.Files, error) {
	return storage.Db.FilesByFingerprint(fingerprint)
}

// The number of files with the specified tag.
func (storage *Storage) FileCountWithTag(tagId uint) (uint, error) {
	return storage.Db.FileCountWithTag(tagId)
}

// Retrieves the set of files with the specified tag.
func (storage *Storage) FilesWithTag(tagId uint) (database.Files, error) {
	return storage.Db.FilesWithTag(tagId)
}

// The number of files with the specified set of tags.
func (storage *Storage) FileCountWithTags(tagIds []uint) (uint, error) {
	if len(tagIds) == 1 {
		fmt.Println("Shortcut as single tag")
		return storage.Db.FileCountWithTag(tagIds[0])
	}

	return storage.Db.FileCountWithTags(tagIds)
}

// Retrieves the set of files with the specified set of tags.
func (storage *Storage) FilesWithTags(includeTagIds, excludeTagIds []uint) (database.Files, error) {
	files, err := storage.Db.FilesWithTags(includeTagIds, excludeTagIds)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve files: %v", err)
	}

	resultFiles := make(database.Files, 0, len(files))
	for _, file := range files {
		if file != nil {
			resultFiles = append(resultFiles, file)
		}
	}

	return resultFiles, nil
}

// Retrieves the set of files that match the specified query.
func (storage *Storage) QueryFiles(expression query.Expression) (database.Files, error) {
	return storage.Db.QueryFiles(expression)
}

// Retrieves the sets of duplicate files within the database.
func (storage *Storage) DuplicateFiles() ([]database.Files, error) {
	return storage.Db.DuplicateFiles()
}

// Adds a file to the database.
func (storage *Storage) AddFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*database.File, error) {
	return storage.Db.InsertFile(path, fingerprint, modTime, size, isDir)
}

// Updates a file in the database.
func (storage *Storage) UpdateFile(fileId uint, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*database.File, error) {
	return storage.Db.UpdateFile(fileId, path, fingerprint, modTime, size, isDir)
}

// Removes a file from the database.
func (storage *Storage) RemoveFile(fileId uint) error {
	return storage.Db.DeleteFile(fileId)
}
