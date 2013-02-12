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
	"fmt"
	"strconv"
)

type FileTag struct {
	FileId uint
	TagId  uint
}

type FileTags []*FileTag

// Determines whether the specified file has the specified tag applied.
func (db *Database) FileTagExists(fileId, tagId uint) (bool, error) {
	sql := `SELECT count(1)
            FROM file_tag
            WHERE file_id = ?1 AND tag_id = ?2`

	rows, err := db.connection.Query(sql, fileId, tagId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	count, err := readCount(rows)
	return count > 0, err
}

// Retrieves the total count of file tags in the database.
func (db *Database) FileTagCount() (uint, error) {
	var sql string

	sql = `SELECT count(1) FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the complete set of file tags.
func (db *Database) FileTags() (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the count of file tags for the specified file.
func (db *Database) FileTagCountByFileId(fileId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1) FROM file_tag WHERE file_id = ?1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of file tags with the specified tag ID.
func (db *Database) FileTagsByTagId(tagId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM file_tag
	        WHERE tag_id = ?1`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of file tags with the specified file ID.
func (db *Database) FileTagsByFileId(fileId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
            FROM file_tag
            WHERE file_id = ?1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Adds a file tag.
func (db *Database) AddFileTag(fileId, tagId uint) (*FileTag, error) {
	sql := `INSERT OR IGNORE INTO file_tag (file_id, tag_id)
            VALUES (?1, ?2)`

	_, err := db.connection.Exec(sql, fileId, tagId)
	if err != nil {
		return nil, err
	}

	return &FileTag{fileId, tagId}, nil
}

// Adds a set of file tags.
func (db *Database) AddFileTags(fileId uint, tagIds []uint) error {
	sql := `INSERT OR IGNORE INTO file_tag (file_id, tag_id)
            VALUES `

	params := make([]interface{}, len(tagIds)+1)
	params[0] = fileId

	for index, tagId := range tagIds {
		params[index+1] = tagId

		if index > 0 {
			sql += ", "
		}

		sql += fmt.Sprintf("(?1, ?%v)", strconv.Itoa(index+2))
	}

	_, err := db.connection.Exec(sql, params...)
	if err != nil {
		return err
	}

	return nil
}

// Removes an file tag.
func (db *Database) DeleteFileTag(fileId, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?1 AND tag_id = ?2`

	result, err := db.connection.Exec(sql, fileId, tagId)
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

// Removes all of the file tags for the specified file.
func (db *Database) DeleteFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the file tags for the specified tag.
func (db *Database) DeleteFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Copies file tags from one tag to another.
func (db *Database) CopyFileTags(sourceTagId uint, destTagId uint) error {
	sql := `INSERT INTO file_tag (file_id, tag_id)
            SELECT file_id, ?2
            FROM file_tag
            WHERE tag_id = ?1`

	_, err := db.connection.Exec(sql, sourceTagId, destTagId)
	if err != nil {
		return err
	}

	return nil
}

// helpers

func readFileTags(rows *sql.Rows, fileTags FileTags) (FileTags, error) {
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		var fileId, tagId uint
		err := rows.Scan(&fileId, &tagId)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, &FileTag{fileId, tagId})
	}

	return fileTags, nil
}
