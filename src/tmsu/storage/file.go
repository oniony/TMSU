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
	"path/filepath"
	"time"
	"tmsu/common/fingerprint"
	"tmsu/entities"
	"tmsu/query"
)

// Retrieves the total number of tracked files.
func (storage *Storage) FileCount() (uint, error) {
	return storage.Db.FileCount()
}

// The complete set of tracked files.
func (storage *Storage) Files() (entities.Files, error) {
	return storage.Db.Files()
}

// Retrieves a specific file.
func (storage *Storage) File(id uint) (*entities.File, error) {
	return storage.Db.File(id)
}

// Retrieves the file with the specified path.
func (storage *Storage) FileByPath(path string) (*entities.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%v: could not retrieve absolute path: %v", path, err)
	}

	return storage.Db.FileByPath(absPath)
}

// Retrieves all files that are under the specified directory.
func (storage *Storage) FilesByDirectory(path string) (entities.Files, error) {
	return storage.Db.FilesByDirectory(path)
}

// Retrieves all file that are under the specified directories.
func (storage *Storage) FilesByDirectories(paths []string) (entities.Files, error) {
	files := make(entities.Files, 0, 100)

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
func (storage *Storage) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	return storage.Db.FilesByFingerprint(fingerprint)
}

// Retrieves the count of files with the specified tags.
func (storage *Storage) FileCountWithTags(tagNames []string) (uint, error) {
	expression := query.HasAll(tagNames)
	return storage.Db.QueryFileCount(expression)
}

// Retrieves the set of files with the specified tags.
func (storage *Storage) FilesWithTags(tagNames []string) (entities.Files, error) {
	expression := query.HasAll(tagNames)
	return storage.Db.QueryFiles(expression)
}

// Retrieves the count of files that match the specified query.
func (storage *Storage) QueryFileCount(expression query.Expression) (uint, error) {
	return storage.Db.QueryFileCount(expression)
}

// Retrieves the set of files that match the specified query.
func (storage *Storage) QueryFiles(expression query.Expression) (entities.Files, error) {
	return storage.Db.QueryFiles(expression)
}

// Retrieves the sets of duplicate files within the database.
func (storage *Storage) DuplicateFiles() ([]entities.Files, error) {
	return storage.Db.DuplicateFiles()
}

// Adds a file to the database.
func (storage *Storage) AddFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	return storage.Db.InsertFile(path, fingerprint, modTime, size, isDir)
}

// Updates a file in the database.
func (storage *Storage) UpdateFile(fileId uint, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	return storage.Db.UpdateFile(fileId, path, fingerprint, modTime, size, isDir)
}

// Deletes a file from the database.
func (storage *Storage) DeleteFile(fileId uint) error {
	return storage.Db.DeleteFile(fileId)
}
