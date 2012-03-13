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
    "errors"
    "strconv"
    "strings"
)

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

func (db Database) Tags() ([]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var id uint
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}

		tags = append(tags, Tag{id, name})
	}

	return tags, nil
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

func (db Database) TagsByFileId(fileId uint) ([]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT tag_id
                FROM file_tag
                WHERE file_id = ?
            )
            ORDER BY name`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var tagId uint
		var tagName string
		err = rows.Scan(&tagId, &tagName)
		if err != nil {
			return nil, err
		}

		tags = append(tags, Tag{tagId, tagName})
	}

	return tags, nil
}

func (db Database) TagsForTags(tagNames []string) ([]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT distinct(tag_id)
                FROM file_tag
                WHERE file_id IN (
                    SELECT file_id
                    FROM file_tag
                    WHERE tag_id IN (
                        SELECT id
                        FROM tag
                        WHERE name IN (` + strings.Repeat("?,", len(tagNames)-1) + `?)
                    )
                    GROUP BY file_id
                    HAVING count(*) = ` + strconv.Itoa(len(tagNames)) + `
                )
            )
            ORDER BY name`

	// convert string array to empty-interface array
	castTagNames := make([]interface{}, len(tagNames))
	for index, tagName := range tagNames {
		castTagNames[index] = tagName
	}

	rows, err := db.connection.Query(sql, castTagNames...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var tagId uint
		var tagName string
		err = rows.Scan(&tagId, &tagName)
		if err != nil {
			return nil, err
		}

		if !db.contains(tagNames, tagName) {
			tags = append(tags, Tag{tagId, tagName})
		}
	}

	return tags, nil
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
