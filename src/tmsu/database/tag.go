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

func (db Database) TagCount() (uint, error) {
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

func (db Database) TagsForTags(tagIds []uint) (Tags, error) {
	files, err := db.FilesWithTags(tagIds, []uint{}, false)
	if err != nil {
		return nil, err
	}

	furtherTags := make(Tags, 0, 10)
	for _, file := range files {
		tags, err := db.TagsByFileId(file.Id)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags {
			if !containsTagId(tagIds, tag.Id) && !containsTag(furtherTags, tag) {
				furtherTags = append(furtherTags, tag)
			}
		}
	}

	return furtherTags, nil
}

func (db Database) AddTag(name string) (*Tag, error) {
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

func (db Database) CopyTag(sourceTagId uint, name string) (*Tag, error) {
	sql := `INSERT INTO tag (name)
            VALUES (?)`

	result, err := db.connection.Exec(sql, name)
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

	destTagId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	sql = `INSERT INTO file_tag (file_id, tag_id)
           SELECT file_id, ?
           FROM file_tag
           WHERE tag_id = ?`

	result, err = db.connection.Exec(sql, destTagId, sourceTagId)
	if err != nil {
		return nil, err
	}

	return &Tag{uint(destTagId), name}, nil
}

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

func containsTagId(items []uint, searchItem uint) bool {
	for _, item := range items {
		if item == searchItem {
			return true
		}
	}

	return false
}

func containsTag(tags Tags, searchTag *Tag) bool {
	for _, tag := range tags {
		if tag.Id == searchTag.Id {
			return true
		}
	}

	return false
}
