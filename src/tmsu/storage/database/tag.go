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

package database

import (
	"database/sql"
	"errors"
	"strings"
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

func (tags Tags) Any(predicate func(*Tag) bool) bool {
	for _, tag := range tags {
		if predicate(tag) {
			return true
		}
	}

	return false
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

	return readCount(rows)
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
func (db Database) Tag(id uint) (*Tag, error) {
	sql := `SELECT id, name
	        FROM tag
	        WHERE id = ?`

	rows, err := db.connection.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTag(rows)
}

// Retrieves a specific tag.
func (db Database) TagByName(name string) (*Tag, error) {
	sql := `SELECT id, name
	        FROM tag
	        WHERE name = ?`

	rows, err := db.connection.Query(sql, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTag(rows)
}

// Retrieves the set of named tags.
func (db Database) TagsByNames(names []string) (Tags, error) {
	if len(names) == 0 {
		return make(Tags, 0), nil
	}

	sql := `SELECT id, name
            FROM tag
            WHERE name IN (?`
	sql += strings.Repeat(",?", len(names)-1)
	sql += ")"

	params := make([]interface{}, len(names))
	for index, name := range names {
		params[index] = name
	}

	result, err := db.connection.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	tags, err := readTags(result, make(Tags, 0, len(names)))
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// Retrieves the set of tags for the specified file.
func (db *Database) TagsByFileId(fileId uint) (Tags, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT tag_id
                FROM file_tag
                WHERE file_id = ?1)
            ORDER BY name`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTags(rows, make(Tags, 0, 10))
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
		return nil, errors.New("expected exactly one row to be affected.")
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
		return nil, errors.New("expected exactly one row to be affected.")
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
	if rowsAffected > 1 {
		return errors.New("expected only one row to be affected.")
	}

	return nil
}

// 

func containsName(tags Tags, name string) bool {
	for _, tag := range tags {
		if tag.Name == name {
			return true
		}
	}

	return false
}

func readTag(rows *sql.Rows) (*Tag, error) {
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

	return &Tag{id, name}, nil
}

func readTags(rows *sql.Rows, tags Tags) (Tags, error) {
	for {
		tag, err := readTag(rows)
		if err != nil {
			return nil, err
		}
		if tag == nil {
			break
		}

		tags = append(tags, tag)
	}

	return tags, nil
}
