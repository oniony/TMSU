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

package database

import (
	"database/sql"
	"errors"
)

type Tag struct {
	Id   uint
	Name string
}

type Tags []*Tag

func (tags Tags) Len() int {
	return len(tags)
}

func (tags Tags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}

func (tags Tags) Less(i, j int) bool {
	return tags[i].Name < tags[j].Name
}

// The number of tags in the database.
func (db *Database) TagCount() (uint, error) {
	sql := `SELECT count(1)
			FROM tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("Could not get tag count.")
	}
	if rows.Err() != nil {
		return 0, err
	}

	var count uint
	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// The set of tags.
func (db Database) Tags() (Tags, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTags(rows, make(Tags, 0, 10))
}

// Retrieves a specific tag.
func (db Database) TagByName(name string) (*Tag, error) {
	sql := `SELECT id
	        FROM tag
	        WHERE name = ?`

	rows, err := db.connection.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, err
	}

	var id uint
	err = rows.Scan(&id)
	if err != nil {
		return nil, err
	}

	return &Tag{id, name}, nil
}

// Adds a tag.
func (db Database) InsertTag(name string) (*Tag, error) {
	sql := `INSERT INTO tag (name)
	        VALUES (?)`

	result, err := db.connection.Exec(sql, name)
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
		return nil, errors.New("Expected exactly one row to be affected.")
	}

	return &Tag{uint(id), name}, nil
}

// Renames a tag.
func (db Database) RenameTag(tagId uint, name string) (*Tag, error) {
	sql := `UPDATE tag
	        SET name = ?
	        WHERE id = ?`

	result, err := db.connection.Exec(sql, name, tagId)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		return nil, errors.New("Expected exactly one row to be affected.")
	}

	return &Tag{tagId, name}, nil
}

// Deletes a tag.
func (db Database) DeleteTag(tagId uint) error {
	sql := `DELETE FROM tag
	        WHERE id = ?`

	result, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("Expected exactly one row to be affected.")
	}

	return nil
}

// 

func readTags(rows *sql.Rows, tags Tags) (Tags, error) {
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		var tagId uint
		var tagName string
		err := rows.Scan(&tagId, &tagName)
		if err != nil {
			return nil, err
		}

		tags = append(tags, &Tag{tagId, tagName})
	}

	return tags, nil
}
