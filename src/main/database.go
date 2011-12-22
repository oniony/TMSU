/*
Copyright 2011 Paul Ruane.

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

package main

import (
    "exp/sql"
    "errors"
    "path/filepath"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	connection *sql.DB
}

func OpenDatabase(path string) (*Database, error) {
	connection, error := sql.Open("sqlite3", path)
	if error != nil { return nil, error }

	database := Database{connection}

	error = database.CreateSchema()
	if error != nil { return nil, error }

	return &database, nil
}

func (this *Database) Close() error {
	return this.connection.Close()
}

func (this Database) TagCount() (uint, error) {
	sql := `SELECT count(1)
			FROM tag`

	rows, error := this.connection.Query(sql)
	if error != nil { return 0, error }
	defer rows.Close()

	if !rows.Next() { return 0, errors.New("Could not get tag count.") }
	if rows.Err() != nil { return 0, error }

	var count uint
	error = rows.Scan(&count)
	if error != nil { return 0, error }

	return count, nil
}

func (this Database) Tags() ([]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
	    if rows.Err() != nil { return nil, error }

		var id uint
		var name string
		error = rows.Scan(&id, &name)
		if error != nil { return nil, error }

		tags = append(tags, Tag{id, name})
	}

	return tags, nil
}

func (this Database) TagByName(name string) (*Tag, error) {
	sql := `SELECT id
	        FROM tag
	        WHERE name = ?`

	rows, error := this.connection.Query(sql, name)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() { return nil, nil }
	if rows.Err() != nil { return nil, error }

	var id uint
	error = rows.Scan(&id)
	if error != nil { return nil, error }

	return &Tag{id, name}, nil
}

func (this Database) TagsByFileId(fileId uint) ([]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                SELECT tag_id
                FROM file_tag
                WHERE file_id = ?
            )
            ORDER BY name`

	rows, error := this.connection.Query(sql, fileId)
	if error != nil { return nil, error }
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var tagId uint
		var tagName string
		error = rows.Scan(&tagId, &tagName)
		if error != nil { return nil, error }

		tags = append(tags, Tag{tagId, tagName})
	}

	return tags, nil
}

func (this Database) TagsForTags(tagNames []string) ([]Tag, error) {
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

	rows, error := this.connection.Query(sql, castTagNames...)
	if error != nil { return nil, error }
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var tagId uint
		var tagName string
		error = rows.Scan(&tagId, &tagName)
		if error != nil { return nil, error }

		if !this.contains(tagNames, tagName) {
			tags = append(tags, Tag{tagId, tagName})
		}
	}

	return tags, nil
}

func (this Database) AddTag(name string) (*Tag, error) {
	sql := `INSERT INTO tag (name)
	        VALUES (?)`

	result, error := this.connection.Exec(sql, name)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return nil, error }
	if rowsAffected != 1 { return nil, errors.New("Expected exactly one row to be affected.") }

	return &Tag{uint(id), name}, nil
}

func (this Database) RenameTag(tagId uint, name string) (*Tag, error) {
	sql := `UPDATE tag
	        SET name = ?
	        WHERE id = ?`

	result, error := this.connection.Exec(sql, name, tagId)
	if error != nil { return nil, error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return nil, error }
	if rowsAffected != 1 { return nil, errors.New("Expected exactly one row to be affected.") }

	return &Tag{tagId, name}, nil
}

func (this Database) DeleteTag(tagId uint) error {
	sql := `DELETE FROM tag
	        WHERE id = ?`

	result, error := this.connection.Exec(sql, tagId)
	if error != nil { return error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return error }
	if rowsAffected != 1 { return errors.New("Expected exactly one row to be affected.") }

	return nil
}

func (this Database) FileCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file`

	rows, error := this.connection.Query(sql)
	if error != nil { return 0, error }
	defer rows.Close()

	if !rows.Next() { return 0, errors.New("Could not get file count.") }
	if rows.Err() != nil { return 0, error }

	var count uint
	error = rows.Scan(&count)
	if error != nil { return 0, error }

	return count, nil
}

func (this Database) Files() ([]File, error) {
	sql := `SELECT id, directory, name, fingerprint
	        FROM file`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		error = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if error != nil { return nil, error }

		files = append(files, File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (this Database) File(id uint) (*File, error) {
	sql := `SELECT directory, name, fingerprint
	        FROM file
	        WHERE id = ?`

	rows, error := this.connection.Query(sql, id)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() { return nil, nil }
	if rows.Err() != nil { return nil, error }

	var directory string
	var name string
	var fingerprint string
	error = rows.Scan(&directory, &name, &fingerprint)
	if error != nil { return nil, error }

	return &File{id, directory, name, fingerprint}, nil
}

func (this Database) FileByPath(path string) (*File, error) {
    directory, name := filepath.Split(path)
    directory = filepath.Clean(directory)

	sql := `SELECT id, fingerprint
	        FROM file
	        WHERE directory = ? AND name = ?`

	rows, error := this.connection.Query(sql, directory, name)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() { return nil, nil }
	if rows.Err() != nil { return nil, error }

	var id uint
	var fingerprint string
	error = rows.Scan(&id, &fingerprint)
	if error != nil { return nil, error }

	return &File{id, directory, name, fingerprint}, nil
}

func (this Database) FilesByDirectory(directory string) ([]File, error) {
	sql := `SELECT id, name, fingerprint
	        FROM file
	        WHERE directory = ?`

	rows, error := this.connection.Query(sql, directory)
	if error != nil { return nil, error }
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileId uint
		var name string
		var fingerprint string
		error = rows.Scan(&fileId, &name, &fingerprint)
		if error != nil { return nil, error }

		files = append(files, File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (this Database) FilesByFingerprint(fingerprint string) ([]File, error) {
	sql := `SELECT id, directory, name
	        FROM file
	        WHERE fingerprint = ?`

	rows, error := this.connection.Query(sql, fingerprint)
	if error != nil { return nil, error }
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileId uint
		var directory string
		var name string
		error = rows.Scan(&fileId, &directory, &name)
		if error != nil { return nil, error }

		files = append(files, File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (this Database) DuplicateFiles() ([][]File, error) {
    sql := `SELECT id, directory, name, fingerprint
            FROM file
            WHERE fingerprint IN (SELECT fingerprint
                                FROM file
                                GROUP BY fingerprint
                                HAVING count(1) > 1)
            ORDER BY fingerprint`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

    fileSets := make([][]File, 0, 10)
    var fileSet []File
    var previousFingerprint string

	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		error = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if error != nil { return nil, error }

	    if fingerprint != previousFingerprint {
	        if fileSet != nil { fileSets = append(fileSets, fileSet) }
            fileSet = make([]File, 0, 10)
            previousFingerprint = fingerprint
        }

		fileSet = append(fileSet, File{fileId, directory, name, fingerprint})
	}

    // ensure last file set is added
    if len(fileSet) > 0 { fileSets = append(fileSets, fileSet) }

	return fileSets, nil
}

func (this Database) AddFile(path string, fingerprint string) (*File, error) {
    directory, name := filepath.Split(path)
    directory = filepath.Clean(directory) //TODO remove when patched

	sql := `INSERT INTO file (directory, name, fingerprint)
	        VALUES (?, ?, ?)`

	result, error := this.connection.Exec(sql, directory, name, fingerprint)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return nil, error }
	if rowsAffected != 1 { return nil, errors.New("Expected exactly one row to be affected.") }

	return &File{uint(id), directory, name, fingerprint}, nil
}

func (this Database) FilesWithTags(tagNames []string) ([]File, error) {
	sql := `SELECT id, directory, name, fingerprint
            FROM file
            WHERE id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (
                    SELECT id
                    FROM tag
                    WHERE name IN (` + strings.Repeat("?,", len(tagNames)-1) + `?)
                )
                GROUP BY file_id
                HAVING count(*) = ` + strconv.Itoa(len(tagNames)) + `
            )`

	// convert string array to empty-interface array
	castTagNames := make([]interface{}, len(tagNames))
	for index, tagName := range tagNames {
		castTagNames[index] = tagName
	}

	rows, error := this.connection.Query(sql, castTagNames...)
	if error != nil { return nil, error }
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		error = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if error != nil { return nil, error }

		files = append(files, File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (this Database) UpdateFileFingerprint(fileId uint, fingerprint string) error {
	sql := `UPDATE file
	        SET fingerprint = ?
	        WHERE id = ?`

	_, error := this.connection.Exec(sql, fingerprint, int(fileId))
	if error != nil { return error }

	return nil
}

func (this Database) RemoveFile(fileId uint) error {
	sql := `DELETE FROM file
	        WHERE id = ?`

	result, error := this.connection.Exec(sql, fileId)
	if error != nil { return error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return error }
	if rowsAffected != 1 { return errors.New("Expected exactly one row to be affected.") }

	return nil
}

func (this Database) FileTagCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file_tag`

	rows, error := this.connection.Query(sql)
	if error != nil { return 0, error }
	defer rows.Close()

	if !rows.Next() { return 0, errors.New("Could not get file-tag count.") }
	if rows.Err() != nil { return 0, error }

	var count uint
	error = rows.Scan(&count)
	if error != nil { return 0, error }

	return count, nil
}

func (this Database) FileTags() ([]FileTag, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileTagId uint
		var fileId uint
		var tagId uint
		error = rows.Scan(&fileTagId, &fileId, &tagId)
		if error != nil { return nil, error }

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
}

func (this Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*FileTag, error) {
	sql := `SELECT id
	        FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	rows, error := this.connection.Query(sql, fileId, tagId)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() { return nil, nil }
	if rows.Err() != nil { return nil, error }

	var fileTagId uint
	error = rows.Scan(&fileTagId)
    if error != nil { return nil, error }

	return &FileTag{fileTagId, fileId, tagId}, nil
}

func (this Database) FileTagsByTagId(tagId uint) ([]FileTag, error) {
	sql := `SELECT id, file_id
	        FROM file_tag
	        WHERE tag_id = ?`

	rows, error := this.connection.Query(sql, tagId)
	if error != nil { return nil, error }
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
	for rows.Next() {
        if rows.Err() != nil { return nil, error }

		var fileTagId uint
		var fileId uint
		error = rows.Scan(&fileTagId, &fileId)
		if error != nil { return nil, error }

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
}

func (this Database) AnyFileTagsForFile(fileId uint) (bool, error) {
    sql := `SELECT 1
            FROM file_tag
            WHERE file_id = ?
            LIMIT 1`

	rows, error := this.connection.Query(sql, fileId)
	if error != nil { return false, error }
	defer rows.Close()

    if rows.Next() { return true, nil }
	if rows.Err() != nil { return false, error }

    return false, nil
}

func (this Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
	sql := `INSERT INTO file_tag (file_id, tag_id)
	        VALUES (?, ?)`

	result, error := this.connection.Exec(sql, fileId, tagId)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return nil, error }
	if rowsAffected != 1 { return nil, errors.New("Expected exactly one row to be affected.") }

	return &FileTag{uint(id), fileId, tagId}, nil
}

func (this Database) RemoveFileTag(fileId uint, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	result, error := this.connection.Exec(sql, fileId, tagId)
	if error != nil { return error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return error }
	if rowsAffected != 1 { return errors.New("Expected exactly one row to be affected.") }

	return nil
}

func (this Database) RemoveFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	_, error := this.connection.Exec(sql, fileId)
	if error != nil { return error }

	return nil
}

func (this Database) RemoveFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	_, error := this.connection.Exec(sql, tagId)
	if error != nil { return error }

	return nil
}

func (this Database) UpdateFileTags(oldTagId uint, newTagId uint) error {
	sql := `UPDATE file_tag
	        SET tag_id = ?
	        WHERE tag_id = ?`

	result, error := this.connection.Exec(sql, newTagId, oldTagId)
	if error != nil { return error }

	rowsAffected, error := result.RowsAffected()
	if error != nil { return error }
	if rowsAffected != 1 { return errors.New("Expected exactly one row to be affected.") }

	return nil
}

func (this Database) CreateSchema() error {
    sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	_, error := this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE TABLE IF NOT EXISTS file (
               id INTEGER PRIMARY KEY,
               directory TEXT NOT NULL,
               name TEXT NOT NULL,
               fingerprint TEXT NOT NULL,
               CONSTRAINT con_file_path UNIQUE (directory, name)
           )`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_path
           ON file(directory, name)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE TABLE IF NOT EXISTS file_tag (
               id INTEGER PRIMARY KEY,
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
               FOREIGN KEY (file_id) REFERENCES file(id),
               FOREIGN KEY (tag_id) REFERENCES tag(id)
           )`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

	return nil
}

func (this Database) contains(list []string, str string) bool {
	for _, current := range list {
		if current == str {
			return true
		}
	}

	return false
}
