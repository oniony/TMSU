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
	"tmsu/storage/database"
)

type Storage struct {
	Db *database.Database
}

func Open() (*Storage, error) {
	db, err := database.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open database: %v", err)
	}

	return &Storage{db}, nil
}

func OpenAt(path string) (*Storage, error) {
	db, err := database.OpenAt(path)
	if err != nil {
		return nil, fmt.Errorf("could not open database at '%v': %v", path, err)
	}

	return &Storage{db}, nil
}

func (storage *Storage) Close() error {
	err := storage.Db.Close()
	if err != nil {
		return fmt.Errorf("could not close database: %v", err)
	}

	return nil
}
