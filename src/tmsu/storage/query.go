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
	"tmsu/entities"
)

// The complete set of queries.
func (storage *Storage) Queries() (entities.Queries, error) {
	return storage.Db.Queries()
}

// Retrievs the specified query.
func (storage *Storage) Query(text string) (*entities.Query, error) {
	return storage.Db.Query(text)
}

// Adds a query to the database.
func (storage *Storage) AddQuery(text string) (*entities.Query, error) {
	return storage.Db.InsertQuery(text)
}

// Removes a query from the database.
func (storage *Storage) DeleteQuery(text string) error {
	return storage.Db.DeleteQuery(text)
}
