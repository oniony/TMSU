// Copyright 2011-2015 Paul Ruane.

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
	"fmt"
	"path/filepath"
	"time"
	"tmsu/common/fingerprint"
	_path "tmsu/common/path"
	"tmsu/entities"
	"tmsu/query"
	"tmsu/storage/database"
)

// Retrieves the total number of tracked files.
func (store *Storage) FileCount(tx *Tx) (uint, error) {
	return database.FileCount(tx.tx)
}

// The complete set of tracked files.
func (store *Storage) Files(tx *Tx, sort string) (entities.Files, error) {
	files, err := database.Files(tx.tx, sort)
	store.absPaths(files)

	return files, err
}

// Retrieves a specific file.
func (store *Storage) File(tx *Tx, id entities.FileId) (*entities.File, error) {
	file, err := database.File(tx.tx, id)
	store.absPath(file)

	return file, err
}

// Retrieves the file with the specified path.
func (store *Storage) FileByPath(tx *Tx, path string) (*entities.File, error) {
	relPath := store.relPath(path)
	file, err := database.FileByPath(tx.tx, relPath)
	store.absPath(file)

	return file, err
}

// Retrieves all files that are under the specified directory.
func (store *Storage) FilesByDirectory(tx *Tx, path string) (entities.Files, error) {
	relPath := store.relPath(path)
	files, err := database.FilesByDirectory(tx.tx, relPath)
	store.absPaths(files)

	return files, err
}

// Retrieves all file that are under the specified directories.
func (store *Storage) FilesByDirectories(tx *Tx, paths []string) (entities.Files, error) {
	files := make(entities.Files, 0, 100)

	for _, path := range paths {
		relPath := store.relPath(path)
		pathFiles, err := database.FilesByDirectory(tx.tx, relPath)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve files for directory: %v", path, err)
		}

		files = append(files, pathFiles...)
	}

	store.absPaths(files)

	return files, nil
}

// Retrieves the number of files with the specified fingerprint.
func (store *Storage) FileCountByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (uint, error) {
	return database.FileCountByFingerprint(tx.tx, fingerprint)
}

// Retrieves the set of files with the specified fingerprint.
func (store *Storage) FilesByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	files, err := database.FilesByFingerprint(tx.tx, fingerprint)
	store.absPaths(files)
	return files, err
}

// Retrieves the set of untagged files.
func (store *Storage) UntaggedFiles(tx *Tx) (entities.Files, error) {
	files, err := database.UntaggedFiles(tx.tx)
	store.absPaths(files)
	return files, err
}

// Retrieves the count of files that match the specified query and matching the specified path.
func (store *Storage) FileCountForQuery(tx *Tx, expression query.Expression, path string, explicitOnly, ignoreCase bool) (uint, error) {
	relPath := store.relPath(path)
	return database.FileCountForQuery(tx.tx, expression, relPath, explicitOnly, ignoreCase)
}

// Retrieves the set of files that match the specified query.
func (store *Storage) FilesForQuery(tx *Tx, expression query.Expression, path string, explicitOnly, ignoreCase bool, sort string) (entities.Files, error) {
	relPath := store.relPath(path)
	files, err := database.FilesForQuery(tx.tx, expression, relPath, explicitOnly, ignoreCase, sort)
	store.absPaths(files)
	return files, err
}

// Retrieves the sets of duplicate files within the database.
func (store *Storage) DuplicateFiles(tx *Tx) ([]entities.Files, error) {
	fileSets, err := database.DuplicateFiles(tx.tx)

	for _, fileSet := range fileSets {
		store.absPaths(fileSet)
	}

	return fileSets, err
}

// Adds a file to the database.
func (store *Storage) AddFile(tx *Tx, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := store.relPath(path)
	file, err := database.InsertFile(tx.tx, relPath, fingerprint, modTime, size, isDir)
	store.absPath(file)

	return file, err
}

// Updates a file in the database.
func (store *Storage) UpdateFile(tx *Tx, fileId entities.FileId, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := store.relPath(path)
	file, err := database.UpdateFile(tx.tx, fileId, relPath, fingerprint, modTime, size, isDir)
	store.absPath(file)

	return file, err
}

// Deletes a file from the database.
func (store *Storage) DeleteFile(tx *Tx, fileId entities.FileId) error {
	return database.DeleteFile(tx.tx, fileId)
}

// Deletes a file if it is untagged
func (store *Storage) DeleteFileIfUntagged(tx *Tx, fileId entities.FileId) error {
	count, err := store.FileTagCountByFileId(tx, fileId, true)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := store.DeleteFile(tx, fileId); err != nil {
			return err
		}
	}

	return nil
}

// Deletes the specified files if they are untagged
func (store *Storage) DeleteUntaggedFiles(tx *Tx, fileIds entities.FileIds) error {
	return database.DeleteUntaggedFiles(tx.tx, fileIds)
}

// unexported

func (store *Storage) relPath(path string) string {
	if path == "" {
		return "" // don't alter empty paths
	}

	return _path.RelTo(path, store.RootPath)
}

func (store *Storage) absPaths(files entities.Files) {
	for _, file := range files {
		store.absPath(file)
	}
}

func (store *Storage) absPath(file *entities.File) {
	if file == nil || file.Directory == "" || file.Directory[0] == filepath.Separator {
		return
	}

	file.Directory = filepath.Join(store.RootPath, file.Directory)
}
