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
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

//TODO handle iteration errors
//TODO handle scan errors
//TODO check rows affected

type Database struct {
	connection *sql.DB
}

func OpenDatabase(path string) (*Database, error) {
	connection, error := sql.Open("sqlite3", path)
	if error != nil { return nil, error }

	database := Database{connection}
	database.CreateSchema()

	return &database, nil
}

func (this *Database) Close() error {
	return this.connection.Close()
}

func (this Database) Tags() (*[]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	tags := make([]Tag, 0, 10)
	for rows.Next() {
		var id uint
		var name string
		rows.Scan(&id, &name)

		tags = append(tags, Tag{id, name})
	}

	return &tags, nil
}

func (this Database) TagByName(name string) (*Tag, error) {
	sql := `SELECT id
	        FROM tag
	        WHERE name = ?`

	rows, error := this.connection.Query(sql, name)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() { return nil, nil }

	var id uint
	rows.Scan(&id)

	return &Tag{id, name}, nil
}

func (this Database) TagsByFileId(fileId uint) (*[]Tag, error) {
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
		var tagId uint
		var tagName string
		rows.Scan(&tagId, &tagName)

		tags = append(tags, Tag{tagId, tagName})
	}

	return &tags, nil
}

func (this Database) TagsForTags(tagNames []string) (*[]Tag, error) {
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
		var tagId uint
		var tagName string
		rows.Scan(&tagId, &tagName)

		if !this.contains(tagNames, tagName) {
			tags = append(tags, Tag{tagId, tagName})
		}
	}

	return &tags, nil
}

func (this Database) AddTag(name string) (*Tag, error) {
	sql := `INSERT INTO tag (name)
	        VALUES (?)`

	result, error := this.connection.Exec(sql, name)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	return &Tag{uint(id), name}, nil
}

func (this Database) RenameTag(tagId uint, name string) (*Tag, error) {
	sql := `UPDATE tag
	        SET name = ?
	        WHERE id = ?`

	_, error := this.connection.Exec(sql, name, tagId)
	if error != nil { return nil, error }

	return &Tag{tagId, name}, nil
}

func (this Database) DeleteTag(tagId uint) error {
	sql := `DELETE FROM tag
	        WHERE id = ?`

	_, error := this.connection.Exec(sql, tagId)
	if error != nil { return error }

	return nil
}

func (this Database) Files() (*[]File, error) {
	sql := `SELECT id, path, fingerprint
	        FROM file`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
		var fileId uint
		var path string
		var fingerprint string
		rows.Scan(&fileId, &path, &fingerprint)

		files = append(files, File{fileId, path, fingerprint})
	}

	return &files, nil
}

func (this Database) File(id uint) (*File, error) {
	sql := `SELECT path, fingerprint
	        FROM file
	        WHERE id = ?`

	rows, error := this.connection.Query(sql, id)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var path string
	var fingerprint string
	rows.Scan(&path, &fingerprint)

	return &File{id, path, fingerprint}, nil
}

func (this Database) FileByPath(path string) (*File, error) {
	sql := `SELECT id, fingerprint
	        FROM file
	        WHERE path = ?`

	rows, error := this.connection.Query(sql, path)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var id uint
	var fingerprint string
	rows.Scan(&id, &fingerprint)

	return &File{id, path, fingerprint}, nil
}

func (this Database) FileByFingerprint(fingerprint string) (*File, error) {
	sql := `SELECT id, path
	        FROM file
	        WHERE fingerprint = ?`

	rows, error := this.connection.Query(sql, fingerprint)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var id uint
	var path string
	rows.Scan(&id, &path)

	return &File{id, path, fingerprint}, nil
}

func (this Database) AddFile(path string, fingerprint string) (*File, error) {
	sql := `INSERT INTO file (path, fingerprint)
	        VALUES (?,?)`

	result, error := this.connection.Exec(sql, path, fingerprint)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	return &File{uint(id), path, fingerprint}, nil
}

func (this Database) FilesWithTags(tagNames []string) (*[]File, error) {
	sql := `SELECT id, path, fingerprint
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
		var fileId uint
		var path string
		var fingerprint string
		rows.Scan(&fileId, &path, &fingerprint)

		files = append(files, File{fileId, path, fingerprint})
	}

	return &files, nil
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

	_, error := this.connection.Exec(sql, fileId)
	if error != nil { return error }

	return nil
}

func (this Database) FileTags() (*[]FileTag, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, error := this.connection.Query(sql)
	if error != nil { return nil, error }
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
	for rows.Next() {
		var fileTagId uint
		var fileId uint
		var tagId uint
		rows.Scan(&fileTagId, &fileId, &tagId)

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return &fileTags, nil
}

func (this Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*FileTag, error) {
	sql := `SELECT id
	        FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	rows, error := this.connection.Query(sql, fileId, tagId)
	if error != nil { return nil, error }
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var fileTagId uint
	rows.Scan(&fileTagId)

	return &FileTag{fileTagId, fileId, tagId}, nil
}

func (this Database) FileTagsByTagId(tagId uint) (*[]FileTag, error) {
	sql := `SELECT id, file_id
	        FROM file_tag
	        WHERE tag_id = ?`

	rows, error := this.connection.Query(sql, tagId)
	if error != nil { return nil, error }
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
	for rows.Next() {
		var fileTagId uint
		var fileId uint
		rows.Scan(&fileTagId, &fileId)

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return &fileTags, nil
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

    return false, nil
}

func (this Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
	sql := `INSERT INTO file_tag (file_id, tag_id)
	        VALUES (?, ?)`

	result, error := this.connection.Exec(sql, fileId, tagId)
	if error != nil { return nil, error }

	id, error := result.LastInsertId()
	if error != nil { return nil, error }

	return &FileTag{uint(id), fileId, tagId}, nil
}

func (this Database) RemoveFileTag(fileId uint, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	_, error := this.connection.Exec(sql, fileId, tagId)
	if error != nil { return error }

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

func (this Database) MigrateFileTags(oldTagId uint, newTagId uint) error {
	sql := `UPDATE file_tag
	        SET tag_id = ?
	        WHERE tag_id = ?`

	_, error := this.connection.Exec(sql, newTagId, oldTagId)
	if error != nil { return error }

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
               path TEXT UNIQUE NOT NULL,
               fingerprint TEXT NOT NULL
           )`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	_, error = this.connection.Exec(sql)
	if error != nil { return error }

    sql = `CREATE INDEX IF NOT EXISTS idx_file_path
           ON file(path)`

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
