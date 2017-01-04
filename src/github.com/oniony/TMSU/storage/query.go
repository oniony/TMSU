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

// The complete set of queries.
func (storage *Storage) Queries(tx *Tx) (entities.Queries, error) {
	return database.Queries(tx.tx)
}

// Retrievs the specified query.
func (storage *Storage) Query(tx *Tx, text string) (*entities.Query, error) {
	return database.Query(tx.tx, text)
}

// Adds a query to the database.
func (storage *Storage) AddQuery(tx *Tx, text string) (*entities.Query, error) {
	return database.InsertQuery(tx.tx, text)
}

// Removes a query from the database.
func (storage *Storage) DeleteQuery(tx *Tx, text string) error {
	return database.DeleteQuery(tx.tx, text)
}
