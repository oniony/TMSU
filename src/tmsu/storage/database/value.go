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

// Retrieves the count of values.
func (db Database) ValueCount() (uint, error) {
	sql := `SELECT count(1)
            FROM value`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return 0, err
	}

	return readCount(rows)
}

// Retrieves a specific value.
func (db Database) Value(id uint) (*entities.Value, error) {
	sql := `SELECT id, name
	        FROM value
	        WHERE id = ?`

	rows, err := db.transaction.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readValue(rows)
}

// Retrieves a specific value by name.
func (db Database) ValueByName(name string) (*entities.Value, error) {
	sql := `SELECT id, name
	        FROM value
	        WHERE name = ?`

	rows, err := db.transaction.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readValue(rows)
}

// Retrieves the set of values for the specified tag.
func (db *Database) ValuesByTagId(tagId uint) (entities.Values, error) {
	sql := `SELECT id, name
            FROM value
            WHERE id IN (
                SELECT value_id
                FROM file_tag
                WHERE tag_id = ?1)
            ORDER BY name`

	rows, err := db.transaction.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readValues(rows, make(entities.Values, 0, 10))
}

// Adds a value.
func (db Database) InsertValue(name string) (*entities.Value, error) {
	sql := `INSERT INTO value (name)
	        VALUES (?)`

	result, err := db.transaction.Exec(sql, name)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
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

	return &entities.Value{uint(id), name}, nil
}

// Deletes a value.
func (db Database) DeleteValue(valueId uint) error {
	sql := `DELETE FROM value
	        WHERE id = ?`

	result, err := db.transaction.Exec(sql, valueId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 1 {
		return errors.New("expected only one row to be affected.")
	}

	return nil
}

// unexported

func readValue(rows *sql.Rows) (*entities.Value, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var id uint
	var name string
	err := rows.Scan(&id, &name)
	if err != nil {
		return nil, err
	}

	return &entities.Value{id, name}, nil
}

func readValues(rows *sql.Rows, values entities.Values) (entities.Values, error) {
	for {
		value, err := readValue(rows)
		if err != nil {
			return nil, err
		}
		if value == nil {
			break
		}

		values = append(values, value)
	}

	return values, nil
}
