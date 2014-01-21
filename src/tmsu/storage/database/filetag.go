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

// Determines whether the specified file has the specified tag applied.
func (db *Database) FileTagExists(fileId, tagId, value_id uint) (bool, error) {
	sql := `SELECT count(1)
            FROM file_tag
            WHERE file_id = ?1 AND tag_id = ?2 AND value_id = ?3`

	rows, err := db.transaction.Query(sql, fileId, tagId, value_id)
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

	sql = `SELECT count(1)
	       FROM file_tag`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the complete set of file tags.
func (db *Database) FileTags() (entities.FileTags, error) {
	sql := `SELECT file_id, tag_id, value_id
	        FROM file_tag`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(entities.FileTags, 0, 10))
}

// Retrieves the count of file tags for the specified file.
func (db *Database) FileTagCountByFileId(fileId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1)
	       FROM file_tag
	       WHERE file_id = ?1`

	rows, err := db.transaction.Query(sql, fileId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the count of file tags for the specified tag.
func (db *Database) FileTagCountByTagId(tagId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1)
	       FROM file_tag
	       WHERE tag_id = ?1`

	rows, err := db.transaction.Query(sql, tagId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of file tags with the specified tag ID.
func (db *Database) FileTagsByTagId(tagId uint) (entities.FileTags, error) {
	sql := `SELECT file_id, tag_id, value_id
	        FROM file_tag
	        WHERE tag_id = ?1`

	rows, err := db.transaction.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(entities.FileTags, 0, 10))
}

// Retrieves the count of file tags for the specified value.
func (db *Database) FileTagCountByValueId(valueId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1)
	       FROM file_tag
	       WHERE value_id = ?1`

	rows, err := db.transaction.Query(sql, valueId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of file tags with the specified value ID.
func (db *Database) FileTagsByValueId(valueId uint) (entities.FileTags, error) {
	sql := `SELECT file_id, tag_id, value_id
	        FROM file_tag
	        WHERE value_id = ?1`

	rows, err := db.transaction.Query(sql, valueId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(entities.FileTags, 0, 10))
}

// Retrieves the set of file tags for the specified file.
func (db *Database) FileTagsByFileId(fileId uint) (entities.FileTags, error) {
	sql := `SELECT file_id, tag_id, value_id
            FROM file_tag
            WHERE file_id = ?1`

	rows, err := db.transaction.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(entities.FileTags, 0, 10))
}

// Adds a file tag.
func (db *Database) AddFileTag(fileId, tagId, valueId uint) (*entities.FileTag, error) {
	sql := `INSERT OR IGNORE INTO file_tag (file_id, tag_id, value_id)
            VALUES (?1, ?2, ?3)`

	_, err := db.transaction.Exec(sql, fileId, tagId, valueId)
	if err != nil {
		return nil, err
	}

	return &entities.FileTag{fileId, tagId, valueId}, nil
}

// Removes a file tag.
func (db *Database) DeleteFileTag(fileId, tagId, valueId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?1 AND tag_id = ?2 AND value_id = ?3`

	result, err := db.transaction.Exec(sql, fileId, tagId, valueId)
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

	_, err := db.transaction.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the file tags for the specified tag.
func (db *Database) DeleteFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	_, err := db.transaction.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Copies file tags from one tag to another.
func (db *Database) CopyFileTags(sourceTagId uint, destTagId uint) error {
	sql := `INSERT INTO file_tag (file_id, tag_id, value_id)
            SELECT file_id, ?2, value_id
            FROM file_tag
            WHERE tag_id = ?1`

	_, err := db.transaction.Exec(sql, sourceTagId, destTagId)
	if err != nil {
		return err
	}

	return nil
}

// helpers

func readFileTags(rows *sql.Rows, fileTags entities.FileTags) (entities.FileTags, error) {
	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		var fileId, tagId, valueId uint
		err := rows.Scan(&fileId, &tagId, &valueId)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, &entities.FileTag{fileId, tagId, valueId})
	}

	return fileTags, nil
}
