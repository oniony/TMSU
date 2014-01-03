/*
Copyright 2011-2014 Paul Ruane.

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

package database

import (
	"database/sql"
	"errors"
	"tmsu/entities"
)

// The complete set of queries.
func (db *Database) Queries() (entities.Queries, error) {
	sql := `SELECT text
	        FROM query
	        ORDER BY text`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readQueries(rows, make(entities.Queries, 0, 10))
}

// Retrieves the specified query.
func (db *Database) Query(text string) (*entities.Query, error) {
	sql := `SELECT 1
            FROM query
            WHERE text = ?`

	rows, err := db.connection.Query(sql, text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readQuery(rows)
}

// Adds a query to the database.
func (db *Database) InsertQuery(text string) (*entities.Query, error) {
	sql := `INSERT INTO query (text)
	        VALUES (?)`

	result, err := db.connection.Exec(sql, text)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		return nil, errors.New("expected exactly one row to be affected.")
	}

	return &entities.Query{text}, nil
}

// Removes a query from the database.
func (db *Database) DeleteQuery(text string) error {
	sql := `DELETE FROM query
	        WHERE text = ?`

	result, err := db.connection.Exec(sql, text)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no such query '" + text + "'.")
	}
	if rowsAffected != 1 {
		return errors.New("expected exactly one row to be affected.")
	}

	return nil
}

//

func readQuery(rows *sql.Rows) (*entities.Query, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var text string
	err := rows.Scan(&text)
	if err != nil {
		return nil, err
	}

	return &entities.Query{text}, nil
}

func readQueries(rows *sql.Rows, queries entities.Queries) (entities.Queries, error) {
	for {
		query, err := readQuery(rows)
		if err != nil {
			return nil, err
		}
		if query == nil {
			break
		}

		queries = append(queries, query)
	}

	return queries, nil
}
