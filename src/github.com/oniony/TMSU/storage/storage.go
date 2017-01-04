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
	"fmt"
	"github.com/oniony/TMSU/common/log"
	"github.com/oniony/TMSU/storage/database"
	"path/filepath"
)

type Storage struct {
	db       *database.Database
	DbPath   string
	RootPath string
}

func CreateAt(path string) error {
	return database.CreateAt(path)
}

func OpenAt(path string) (*Storage, error) {
	db, err := database.OpenAt(path)
	if err != nil {
		return nil, err
	}

	rootPath, err := determineRootPath(path)
	if err != nil {
		return nil, err
	}

	log.Infof(2, "files are stored relative to root path '%v'", rootPath)

	return &Storage{db, path, rootPath}, nil
}

func (storage *Storage) Begin() (*Tx, error) {
	tx, err := storage.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{tx}, nil
}

func (storage *Storage) Close() error {
	if storage.db == nil {
		return nil
	}

	err := storage.db.Close()
	if err != nil {
		return fmt.Errorf("could not close database: %v", err)
	}

	storage.db = nil

	return nil
}

type Tx struct {
	tx *database.Tx
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
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
