/*
Copyright 2011-2015 Paul Ruane.

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
	"tmsu/common/log"
	"tmsu/storage/database"
)

type Storage struct {
	Db       *database.Database
	RootPath string
}

func OpenAt(path string) (*Storage, error) {
	db, err := database.OpenAt(path)
	if err != nil {
		return nil, fmt.Errorf("could not open database at '%v': %v", path, err)
	}

	rootPath, err := determineRootPath(path)
	if err != nil {
		return nil, err
	}

	log.Infof(2, "files are stored relative to root path '%v'", rootPath)

	return &Storage{db, rootPath}, nil
}

func (storage *Storage) Begin() error {
	return storage.Db.Begin()
}

func (storage *Storage) Commit() error {
	return storage.Db.Commit()
}

func (storage *Storage) Rollback() error {
	return storage.Db.Rollback()
}

func (storage *Storage) Close() error {
	err := storage.Db.Close()
	if err != nil {
		return fmt.Errorf("could not close database: %v", err)
	}

	return nil
}

// unexported

func determineRootPath(dbPath string) (string, error) {
	absDbPath, err := filepath.Abs(dbPath)
	if err != nil {
		return "", AbsolutePathResolutionError{dbPath, err}
	}

	absDbDirPath := filepath.Dir(absDbPath)
	if filepath.Base(absDbDirPath) == ".tmsu" {
		return filepath.Dir(absDbDirPath), nil
	}

	return string(filepath.Separator), nil //TODO Windows
}
