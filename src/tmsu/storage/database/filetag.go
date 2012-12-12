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

type FileTag struct {
	Id     uint
	FileId uint
	TagId  uint
}

type FileTags []*FileTag

// Retrieves the set of files with the specified tag.
func (db *Database) FilesWithTag(tagId uint) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time
            FROM file
            WHERE id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id = ?)`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

// Retrieves the total count of file tags in the database.
func (db *Database) FileTagCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file_tag`

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
func (db *Database) FileTags() (FileTags, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the set of tags for the specified file.
func (db *Database) TagsByFileId(fileId uint) (Tags, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT tag_id
                FROM file_tag
                WHERE file_id = ?
            )`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readTags(rows, make(Tags, 0, 10))
}

func (db *Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*FileTag, error) {
	sql := `SELECT id
	        FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	rows, err := db.connection.Query(sql, fileId, tagId)
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

	var fileTagId uint
	err = rows.Scan(&fileTagId)
	if err != nil {
		return nil, err
	}

	return &FileTag{fileTagId, fileId, tagId}, nil
}

// Retrieves the file tag with the specified file ID and tag ID.
func (db *Database) FileTagsByTagId(tagId uint) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag
	        WHERE tag_id = ?`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Retrieves the file tags with the specified file ID.
func (db *Database) FileTagsByFileId(fileId uint) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id
            FROM file_tag
            WHERE file_id = ?`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

// Adds a file tag.
func (db *Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
	sql := `INSERT INTO file_tag (file_id, tag_id)
	        VALUES (?, ?)`

	result, err := db.connection.Exec(sql, fileId, tagId)
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

	return &FileTag{uint(id), fileId, tagId}, nil
}

// Removes a file tag.
func (db *Database) RemoveFileTag(fileId uint, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

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
func (db *Database) RemoveFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

// Removes all of the file tags for the specified tag.
func (db *Database) RemoveFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

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
	sql := `INSERT INTO file_tag (file_id, tag_id)
            SELECT file_id, ?
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

		var fileTagId uint
		var fileId uint
		var tagId uint
		err := rows.Scan(&fileTagId, &fileId, &tagId)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, &FileTag{fileTagId, fileId, tagId})
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
