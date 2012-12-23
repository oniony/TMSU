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
	"strings"
)

type FileTag struct {
	Id       uint
	FileId   uint
	TagId    uint
	Explicit bool
	Implicit bool
}

type FileTags []*FileTag

// Retrieves the set of files with the specified tag.
func (db *Database) FilesWithTag(tagId uint, explicitOnly bool) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time
            FROM file
            WHERE id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	sql += `    )
            ORDER BY directory, name`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

func (db *Database) FilesWithTags(tagIds []uint, explicitOnly bool) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size
            FROM file
            WHERE id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (?`

	sql += strings.Repeat(",?", len(tagIds)-1) + ")"

	if explicitOnly {
		sql += " AND explicit"
	}

	sql += `    GROUP BY file_id
                HAVING count(tag_id) == ?)
            ORDER BY directory, name`

	params := make([]interface{}, len(tagIds)+1)
	for index, tagId := range tagIds {
		params[index] = interface{}(tagId)
	}
	params[len(tagIds)] = len(tagIds)

	rows, err := db.connection.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

// Retrieves the total count of file tags in the database.
func (db *Database) FileTagCount(explicitOnly bool) (uint, error) {
	sql := `SELECT count(1)
			FROM file_tag`

	if explicitOnly {
		sql += " WHERE explicit"
	}

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("Could not get file-tag count.")
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

// Retrieves the complete set of file tags.
func (db *Database) FileTags(explicitOnly bool) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id, explicit, implicit
	        FROM file_tag`

	if explicitOnly {
		sql += " WHERE explicit"
	}

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the count of file tags for the specified file.
func (db *Database) FileTagCountByFileId(fileId uint, explicitOnly bool) (uint, error) {
	sql := `SELECT count(1)
            FROM file_tag
            WHERE file_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("Could not get file-tag count.")
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

// Retrieves the set of tags for the specified file.
func (db *Database) TagsByFileId(fileId uint, explicitOnly bool) (Tags, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT tag_id
                FROM file_tag
                WHERE file_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	sql += `    )
            ORDER BY name`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTags(rows, make(Tags, 0, 10))
}

// Retrieves the specified file tag
func (db *Database) FileTag(fileTagId uint) (*FileTag, error) {
	sql := `SELECT id, file_id, tag_id, explicit, implicit
	        FROM file_tag
	        WHERE id = ?`

	rows, err := db.connection.Query(sql, fileTagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags, err := readFileTags(rows, make(FileTags, 0, 1))
	if err != nil {
		return nil, err
	}
	if len(fileTags) == 0 {
		return nil, nil
	}

	return fileTags[0], nil
}

// Retrieves the file tag with the specified file ID and tag ID.
func (db *Database) FileTagByFileIdAndTagId(fileId, tagId uint) (*FileTag, error) {
	sql := `SELECT id, file_id, tag_id, explicit, implicit
	        FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	rows, err := db.connection.Query(sql, fileId, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags, err := readFileTags(rows, make(FileTags, 0, 1))
	if err != nil {
		return nil, err
	}
	if len(fileTags) == 0 {
		return nil, nil
	}

	return fileTags[0], nil
}

// Retrieves the set of file tags with the specified tag ID.
func (db *Database) FileTagsByTagId(tagId uint, explicitOnly bool) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id, explicit, implicit
	        FROM file_tag
	        WHERE tag_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of file tags with the specified file ID.
func (db *Database) FileTagsByFileId(fileId uint, explicitOnly bool) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id, explicit, implicit
            FROM file_tag
            WHERE file_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Adds a file tag.
func (db *Database) InsertFileTag(fileId uint, tagId uint, explicit bool, implicit bool) (*FileTag, error) {
	sql := `INSERT INTO file_tag (file_id, tag_id, explicit, implicit)
	        VALUES (?, ?, ?, ?)`

	result, err := db.connection.Exec(sql, fileId, tagId, explicit, implicit)
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

	return &FileTag{uint(id), fileId, tagId, explicit, implicit}, nil
}

func (db *Database) UpdateFileTag(fileTagId, fileId, tagId uint, explicit bool, implicit bool) (*FileTag, error) {
	sql := `UPDATE file_tag
            SET file_id = ?,
                tag_id = ?,
                explicit = ?,
                implicit = ?
            WHERE id = ?`

	result, err := db.connection.Exec(sql, fileId, tagId, explicit, implicit, fileTagId)
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

	return &FileTag{fileTagId, fileId, tagId, explicit, implicit}, nil
}

// Removes a file tag.
func (db *Database) DeleteFileTag(fileTagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE id = ?`

	result, err := db.connection.Exec(sql, fileTagId)
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

// Remove a file tag by file and tag ID.
func (db *Database) DeleteFileTagByFileAndTagId(fileId uint, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ? AND tag_id = ?`

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

// Removes all of the file tags for the specified file.
func (db *Database) DeleteFileTagsByFileId(fileId uint, explicitOnly bool) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the file tags for the specified tag.
func (db *Database) DeleteFileTagsByTagId(tagId uint, explicitOnly bool) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	if explicitOnly {
		sql += " AND explicit"
	}

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Updates file tags to a new tag.
func (db *Database) UpdateFileTags(oldTagId uint, newTagId uint) error {
	sql := `UPDATE file_tag
	        SET tag_id = ?
	        WHERE tag_id = ?`

	result, err := db.connection.Exec(sql, newTagId, oldTagId)
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

// Copies file tags from one tag to another.
func (db *Database) CopyFileTags(sourceTagId uint, destTagId uint) error {
	sql := `INSERT INTO file_tag (file_id, tag_id, explicit, implicit)
            SELECT file_id, ?, explicit, implicit
            FROM file_tag
            WHERE tag_id = ?`

	_, err := db.connection.Exec(sql, destTagId, sourceTagId)
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

		var fileTagId, fileId, tagId uint
		var explicit, implicit bool
		err := rows.Scan(&fileTagId, &fileId, &tagId, &explicit, &implicit)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, &FileTag{fileTagId, fileId, tagId, explicit, implicit})
	}

	return fileTags, nil
}

func contains(files Files, searchFile *File) bool {
	for _, file := range files {
		if file.Path() == searchFile.Path() {
			return true
		}
	}

	return false
}

func uniq(tags Tags) Tags {
	uniqueTags := make(Tags, 0, len(tags))

	var previousTagName string = ""
	for _, tag := range tags {
		if tag.Name == previousTagName {
			continue
		}

		uniqueTags = append(uniqueTags, tag)
		previousTagName = tag.Name
	}

	return uniqueTags
}
