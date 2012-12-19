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
	"time"
	"tmsu/fingerprint"
	"tmsu/storage/database"
)

// Retrieves the total number of tracked files.
func (storage *Storage) FileCount() (uint, error) {
	return storage.Db.FileCount()
}

// The complete set of tracke files.
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

// Retrieves the number of files with the specified fingerprint.
func (storage *Storage) FileCountByFingerprint(fingerprint fingerprint.Fingerprint) (uint, error) {
	return storage.Db.FileCountByFingerprint(fingerprint)
}

// Retrieves the set of files with the specified fingerprint.
func (storage *Storage) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (database.Files, error) {
	return storage.Db.FilesByFingerprint(fingerprint)
}

// Retrieves the sets of duplicate files within the database.
func (storage *Storage) DuplicateFiles() ([]database.Files, error) {
	return storage.Db.DuplicateFiles()
}

// Adds a file to the database.
func (storage *Storage) AddFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time) (*database.File, error) {
	return storage.Db.InsertFile(path, fingerprint, modTime)
}

// Updates a file in the database.
func (storage *Storage) UpdateFile(fileId uint, path string, fingerprint fingerprint.Fingerprint, modTime time.Time) error {
	return storage.Db.UpdateFile(fileId, path, fingerprint, modTime)
}

// Removes a file from the database.
func (storage *Storage) RemoveFile(fileId uint) error {
	file, err := storage.File(fileId)
	if err != nil {
		return err
	}

	err = storage.Db.DeleteFile(fileId)
	if err != nil {
		return err
	}

	files, err := storage.Db.FilesByDirectory(file.Path())
	if err != nil {
		return err
	}

	for _, file := range files {
		filetags, err := storage.Db.FileTagsByFileId(file.Id, false)
		if err != nil {
			return err
		}

		// remove only untagged descendents
		if len(filetags) == 0 {
			err = storage.Db.DeleteFile(fileId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
