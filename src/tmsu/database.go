// Copyright 2011 Paul Ruane. All rights reserved.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"gosqlite.googlecode.com/hg/sqlite"
)

type Database struct {
	connection *sqlite.Conn
}

func OpenDatabase(path string) (*Database, error) {
	connection, error := sqlite.Open(path)
	if error != nil {
		fmt.Fprintf(os.Stderr, "Could not open database: %v.", error)
		return nil, error
	}

	database := Database{connection}
	database.CreateSchema()

	return &database, nil
}

func (this *Database) Close() {
	this.connection.Close()
}

// tags

func (this Database) Tags() (*[]Tag, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	tags := make([]Tag, 0, 10)
	for statement.Next() {
		var id int 
		var name string
		statement.Scan(&id, &name)

		tags = append(tags, Tag{uint(id), name})
	}

	return &tags, nil
}

func (this Database) TagByName(name string) (*Tag, error) {
	sql := `SELECT id
	        FROM tag
	        WHERE name = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(name)
	if error != nil {
		return nil, error
	}

	if !statement.Next() {
		return nil, nil
	}

	var id int
	statement.Scan(&id)

	return &Tag{uint(id), name}, nil
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

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId))
	if error != nil {
		return nil, error
	}

	tags := make([]Tag, 0, 10)
	for statement.Next() {
		var tagId int
		var tagName string
		statement.Scan(&tagId, &tagName)

		tags = append(tags, Tag{uint(tagId), tagName})
	}

	return &tags, error
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

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	// convert string array to empty-interface array
	castTagNames := make([]interface{}, len(tagNames))
	for index, tagName := range tagNames {
		castTagNames[index] = tagName
	}

	error = statement.Exec(castTagNames...)
	if error != nil {
		return nil, error
	}

	tags := make([]Tag, 0, 10)
	for statement.Next() {
		var tagId int
		var tagName string
		statement.Scan(&tagId, &tagName)

		if !this.contains(tagNames, tagName) {
			tags = append(tags, Tag{uint(tagId), tagName})
		}
	}

	return &tags, nil
}

func (this Database) AddTag(name string) (*Tag, error) {
	sql := `INSERT INTO tag (name)
	        VALUES (?)`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(name)
	if error != nil {
		return nil, error
	}
	statement.Next()

	id := this.connection.LastInsertRowId()

	return &Tag{uint(id), name}, nil
}

func (this Database) RenameTag(tagId uint, name string) (*Tag, error) {
	sql := `UPDATE tag
	        SET name = ?
	        WHERE id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(name, int(tagId))
	if error != nil {
		return nil, error
	}
	statement.Next()

	return &Tag{tagId, name}, nil
}

func (this Database) DeleteTag(tagId uint) error {
	sql := `DELETE FROM tag
	        WHERE id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(tagId))
	if error != nil {
		return error
	}
	statement.Next()

	return nil
}

// files

func (this Database) Files() (*[]File, error) {
	sql := `SELECT id, path, fingerprint
	        FROM file`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec()
	if error != nil {
		return nil, error
	}

	files := make([]File, 0, 10)
	for statement.Next() {
		var fileId int
		var path string
		var fingerprint string
		statement.Scan(&fileId, &path, &fingerprint)

		files = append(files, File{uint(fileId), path, fingerprint})
	}

	return &files, nil
}

func (this Database) File(id uint) (*File, error) {
	sql := `SELECT path, fingerprint
	        FROM file
	        WHERE id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(id))
	if error != nil {
		return nil, error
	}
	if !statement.Next() {
		return nil, nil
	}

	var path string
	var fingerprint string
	statement.Scan(&path, &fingerprint)

	return &File{id, path, fingerprint}, nil
}

func (this Database) FileByPath(path string) (*File, error) {
	sql := `SELECT id, fingerprint
	        FROM file
	        WHERE path = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(path)
	if error != nil {
		return nil, error
	}
	if !statement.Next() {
		return nil, nil
	}

	var id int
	var fingerprint string
	statement.Scan(&id, &fingerprint)

	return &File{uint(id), path, fingerprint}, nil
}

func (this Database) FileByFingerprint(fingerprint string) (*File, error) {
	sql := `SELECT id, path
	        FROM file
	        WHERE fingerprint = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(fingerprint)
	if error != nil {
		return nil, error
	}
	if !statement.Next() {
		return nil, nil
	}

	var id int
	var path string
	statement.Scan(&id, &path)

	return &File{uint(id), path, fingerprint}, nil
}

func (this Database) AddFile(path string, fingerprint string) (*File, error) {
	sql := `INSERT INTO file (path, fingerprint)
	        VALUES (?,?)`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(path, fingerprint)
	if error != nil {
		return nil, error
	}
	statement.Next()

	id := this.connection.LastInsertRowId()

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

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	// convert string array to empty-interface array
	castTagNames := make([]interface{}, len(tagNames))
	for index, tagName := range tagNames {
		castTagNames[index] = tagName
	}

	error = statement.Exec(castTagNames...)
	if error != nil {
		return nil, error
	}

	files := make([]File, 0, 10)
	for statement.Next() {
		var fileId int
		var path string
		var fingerprint string
		statement.Scan(&fileId, &path, &fingerprint)

		files = append(files, File{uint(fileId), path, fingerprint})
	}

	return &files, nil
}

func (this Database) UpdateFileFingerprint(fileId uint, fingerprint string) error {
	sql := `UPDATE file
	        SET fingerprint = ?
	        WHERE id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(fingerprint, int(fileId))
	if error != nil {
		return error
	}
	statement.Next()

	return nil
}

func (this Database) RemoveFile(fileId uint) error {
	sql := `DELETE FROM file
	        WHERE id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId))
	if error != nil {
		return error
	}
	statement.Next()

	return nil
}

func (this Database) FileTags() (*[]FileTag, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec()
	if error != nil {
		return nil, error
	}

	fileTags := make([]FileTag, 0, 10)
	for statement.Next() {
		var fileTagId int
		var fileId int
		var tagId int
		statement.Scan(&fileTagId, &fileId, &tagId)

		fileTags = append(fileTags, FileTag{uint(fileTagId), uint(fileId), uint(tagId)})
	}

	return &fileTags, nil
}

func (this Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*FileTag, error) {
	sql := `SELECT id
	        FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId), int(tagId))
	if error != nil {
		return nil, error
	}
	if !statement.Next() {
		return nil, nil
	}

	var fileTagId int
	statement.Scan(&fileTagId)

	return &FileTag{uint(fileTagId), fileId, tagId}, nil
}

func (this Database) FileTagsByTagId(tagId uint) (*[]FileTag, error) {
	sql := `SELECT id, file_id
	        FROM file_tag
	        WHERE tag_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(tagId))
	if error != nil {
		return nil, error
	}

	fileTags := make([]FileTag, 0, 10)
	for statement.Next() {
		var fileTagId int
		var fileId int
		statement.Scan(&fileTagId, &fileId)

		fileTags = append(fileTags, FileTag{uint(fileTagId), uint(fileId), tagId})
	}

	return &fileTags, nil
}

func (this Database) AnyFileTagsForFile(fileId uint) (bool, error) {
    sql := `SELECT 1
            FROM file_tag
            WHERE file_id = ?
            LIMIT 1`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return false, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId))
	if error != nil {
		return false, error
	}

    return statement.Next(), nil
}

func (this Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
	sql := `INSERT INTO file_tag (file_id, tag_id)
	        VALUES (?, ?)`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return nil, error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId), int(tagId))
	if error != nil {
		return nil, error
	}

	if !statement.Next() {
		return nil, nil
	}

	id := this.connection.LastInsertRowId()

	return &FileTag{uint(id), fileId, tagId}, nil
}

func (this Database) RemoveFileTag(fileId uint, tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?
	        AND tag_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId), int(tagId))
	if error != nil {
		return error
	}

	statement.Next()

	return nil
}

func (this Database) RemoveFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(fileId))
	if error != nil {
		return error
	}

	statement.Next()

	return nil
}

func (this Database) RemoveFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(tagId))
	if error != nil {
		return error
	}

	statement.Next()

	return nil
}

func (this Database) MigrateFileTags(oldTagId uint, newTagId uint) error {
	sql := `UPDATE file_tag
	        SET tag_id = ?
	        WHERE tag_id = ?`

	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec(int(newTagId), int(oldTagId))
	if error != nil {
		return error
	}

	statement.Next()

	return nil
}

func (this Database) CreateSchema() error {
    sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	error := this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE TABLE IF NOT EXISTS file (
               id INTEGER PRIMARY KEY,
               path TEXT UNIQUE NOT NULL,
               fingerprint TEXT NOT NULL
           )`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE INDEX IF NOT EXISTS idx_file_path
           ON file(path)`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE TABLE IF NOT EXISTS file_tag (
               id INTEGER PRIMARY KEY,
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
               FOREIGN KEY (file_id) REFERENCES file(id),
               FOREIGN KEY (tag_id) REFERENCES tag(id)
           )`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	error = this.exec(sql)
	if error != nil {
		return error
	}

    sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	error = this.exec(sql)
	if error != nil {
		return error
	}

	return nil
}

// private

func (this Database) contains(list []string, str string) bool {
	for _, current := range list {
		if current == str {
			return true
		}
	}

	return false
}

func (this *Database) exec(sql string) error {
	statement, error := this.connection.Prepare(sql)
	if error != nil {
		return error
	}
	defer statement.Finalize()

	error = statement.Exec()
	if error != nil {
		return error
	}

	statement.Next()

	return nil
}

