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
            FROM explicit_file_tag
            WHERE file_id = ?1 AND tag_id = ?2
            UNION SELECT count(1)
                  FROM implicit_file_tag
                  WHERE file_id = ?1 AND tag_id = ?2`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	count, err := readCount(rows)
	return count > 0, err
}

// Determines whether the specified file has the specified explicit tag applied.
func (db *Database) ExplicitFileTagExists(fileId, tagId uint) (bool, error) {
	sql := `SELECT count(1)
            FROM explicit_file_tag
            WHERE file_id = ?1 AND tag_id = ?2`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	count, err := readCount(rows)
	return count > 0, err
}

// Determines whether the specified file has the specified implicit tag applied.
func (db *Database) ImplicitFileTagExists(fileId, tagId uint) (bool, error) {
	sql := `SELECT count(1)
            FROM implicit_file_tag
            WHERE file_id = ?1 AND tag_id = ?2`

	rows, err := db.connection.Query(sql)
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

	sql = `SELECT (SELECT count(1) FROM explicit_file_tag) +
                  (SELECT count(1) FROM implicit_file_tag)`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the total count of explicit file tags in the database.
func (db *Database) ExplicitFileTagCount() (uint, error) {
	var sql string

	sql = `SELECT count(1)
           FROM explicit_file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the total count of implicit file tags in the database.
func (db *Database) ImplicitFileTagCount() (uint, error) {
	var sql string

	sql = `SELECT count(1)
           FROM implicit_file_tag`

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
	        FROM explicit_file_tag
            UNION SELECT file_id, tag_id
		          FROM implicit_file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the complete set of file tags.
func (db *Database) ExplicitFileTags() (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM explicit_file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the complete set of file tags.
func (db *Database) ImplicitFileTags() (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM explicit_file_tag`

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

	sql = `SELECT (SELECT count(1) FROM explicit_file_tag WHERE file_id = ?1) +
                  (SELECT count(1) FROM implicit_file_tag WHERE file_id = ?1)`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the count of explicit file tags for the specified file.
func (db *Database) ExplicitFileTagCountByFileId(fileId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1)
           FROM explicit_file_tag
           WHERE file_id = ?`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the count of implicit file tags for the specified file.
func (db *Database) ImplicitFileTagCountByFileId(fileId uint) (uint, error) {
	var sql string

	sql = `SELECT count(1)
           FROM implicit_file_tag
           WHERE file_id = ?`

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
	        FROM explicit_file_tag
	        WHERE tag_id = ?1
            UNION
		    SELECT file_id, tag_id
		    FROM implicit_file_tag
		    WHERE tag_id = ?1`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of explicit file tags with the specified tag ID.
func (db *Database) ExplicitFileTagsByTagId(tagId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM explicit_file_tag
	        WHERE tag_id = ?1`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of implicit file tags with the specified tag ID.
func (db *Database) ImplicitFileTagsByTagId(tagId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
	        FROM implicit_file_tag
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
            FROM explicit_file_tag
            WHERE file_id = ?1
		    UNION
		    SELECT file_id, tag_id
		    FROM implicit_file_tag
		    WHERE file_id = ?1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of explicit file tags with the specified file ID.
func (db *Database) ExplicitFileTagsByFileId(fileId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
            FROM explicit_file_tag
            WHERE file_id = ?1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of implicit file tags with the specified file ID.
func (db *Database) ImplicitFileTagsByFileId(fileId uint) (FileTags, error) {
	sql := `SELECT file_id, tag_id
            FROM implicit_file_tag
            WHERE file_id = ?1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Adds an explicit file tag.
func (db *Database) AddExplicitFileTag(fileId, tagId uint) (*FileTag, error) {
	sql := `INSERT OR IGNORE INTO explicit_file_tag (file_id, tag_id)
            VALUES (?1, ?2)`

	_, err := db.connection.Exec(sql, fileId, tagId)
	if err != nil {
		return nil, err
	}

	return &FileTag{fileId, tagId}, nil
}

// Adds an implicit file tag.
func (db *Database) AddImplicitFileTag(fileId, tagId uint) (*FileTag, error) {
	sql := `INSERT OR IGNORE INTO implicit_file_tag (file_id, tag_id)
            VALUES (?1, ?2)`

	_, err := db.connection.Exec(sql, fileId, tagId)
	if err != nil {
		return nil, err
	}

	return &FileTag{fileId, tagId}, nil
}

// Adds a set of explicit file tags.
func (db *Database) AddExplicitFileTags(fileId uint, tagIds []uint) error {
	sql := `INSERT OR IGNORE INTO explicit_file_tag (file_id, tag_id)
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

// Adds a set of implicit file tags.
func (db *Database) AddImplicitFileTags(fileId uint, tagIds []uint) error {
	sql := `INSERT OR IGNORE INTO implicit_file_tag (file_id, tag_id)
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

// Removes an explicit file tag.
func (db *Database) DeleteExplicitFileTag(fileId, tagId uint) error {
	sql := `DELETE FROM explicit_file_tag
	        WHERE file_id = ?1 AND tag_id = ?2`

	result, err := db.connection.Exec(sql, fileId, tagId)
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

// Removes an implicit file tag.
func (db *Database) DeleteImplicitFileTag(fileId, tagId uint) error {
	sql := `DELETE FROM implicit_file_tag
	        WHERE file_id = ?1 AND tag_id = ?2`

	result, err := db.connection.Exec(sql, fileId, tagId)
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

// Removes all of the explicit file tags for the specified file.
func (db *Database) DeleteExplicitFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM explicit_file_tag
	        WHERE file_id = ?`

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the implicit file tags for the specified file.
func (db *Database) DeleteImplicitFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM implicit_file_tag
	        WHERE file_id = ?`

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the explicit file tags for the specified tag.
func (db *Database) DeleteExplicitFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM explicit_file_tag
	        WHERE tag_id = ?`

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the implicit file tags for the specified tag.
func (db *Database) DeleteImplicitFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM implicit_file_tag
	        WHERE tag_id = ?`

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Copies explicit file tags from one tag to another.
func (db *Database) CopyExplicitFileTags(sourceTagId uint, destTagId uint) error {
	sql := `INSERT INTO explicit_file_tag (file_id, tag_id)
            SELECT file_id, ?2
            FROM explicit_file_tag
            WHERE tag_id = ?1`

	_, err := db.connection.Exec(sql, sourceTagId, destTagId)
	if err != nil {
		return err
	}

	return nil
}

// Copies implicit file tags from one tag to another.
func (db *Database) CopyImplicitFileTags(sourceTagId uint, destTagId uint) error {
	sql := `INSERT INTO implicit_file_tag (file_id, tag_id)
	        SELECT file_id, ?2
	        FROM implicit_file_tag
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
