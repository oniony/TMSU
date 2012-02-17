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

package database

import (
	"errors"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tmsu/core"
	"tmsu/entities"
)

type Database struct {
	connection *sql.DB
}

func OpenDatabase() (*Database, error) {
	config, err := core.GetSelectedDatabaseConfig()
	if err != nil {
		return nil, err
	}
	if config == nil {
		config, err = core.GetDefaultDatabaseConfig()
		if err != nil {
			return nil, errors.New("Could not retrieve default database configuration: " + err.Error())
		}

		// attempt to create default database directory
		dir := filepath.Dir(config.DatabasePath)
		os.MkdirAll(dir, os.ModeDir|0755)
	}

	return OpenDatabaseAt(config.DatabasePath)
}

func OpenDatabaseAt(path string) (*Database, error) {
	connection, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.New("Could not open database: " + err.Error())
	}

	database := Database{connection}

	err = database.CreateSchema()
	if err != nil {
		return nil, errors.New("Could not create database schema: " + err.Error())
	}

	return &database, nil
}

func (db *Database) Close() error {
	return db.connection.Close()
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

func (db Database) Tags() ([]entities.Tag, error) {
	sql := `SELECT id, name
            FROM tag
            ORDER BY name`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]entities.Tag, 0, 10)
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

		tags = append(tags, entities.Tag{id, name})
	}

	return tags, nil
}

func (db Database) TagByName(name string) (*entities.Tag, error) {
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

	return &entities.Tag{id, name}, nil
}

func (db Database) TagsByFileId(fileId uint) ([]entities.Tag, error) {
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

	tags := make([]entities.Tag, 0, 10)
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

		tags = append(tags, entities.Tag{tagId, tagName})
	}

	return tags, nil
}

func (db Database) TagsForTags(tagNames []string) ([]entities.Tag, error) {
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

	tags := make([]entities.Tag, 0, 10)
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
			tags = append(tags, entities.Tag{tagId, tagName})
		}
	}

	return tags, nil
}

func (db Database) AddTag(name string) (*entities.Tag, error) {
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

	return &entities.Tag{uint(id), name}, nil
}

func (db Database) RenameTag(tagId uint, name string) (*entities.Tag, error) {
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

	return &entities.Tag{tagId, name}, nil
}

func (db Database) CopyTag(sourceTagId uint, name string) (*entities.Tag, error) {
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

	return &entities.Tag{uint(destTagId), name}, nil
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

func (db Database) FileCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("Could not get file count.")
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

func (db Database) Files() ([]entities.File, error) {
	sql := `SELECT id, directory, name, fingerprint
	        FROM file`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]entities.File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		err = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if err != nil {
			return nil, err
		}

		files = append(files, entities.File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (db Database) File(id uint) (*entities.File, error) {
	sql := `SELECT directory, name, fingerprint
	        FROM file
	        WHERE id = ?`

	rows, err := db.connection.Query(sql, id)
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

	var directory string
	var name string
	var fingerprint string
	err = rows.Scan(&directory, &name, &fingerprint)
	if err != nil {
		return nil, err
	}

	return &entities.File{id, directory, name, fingerprint}, nil
}

func (db Database) FileByPath(path string) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `SELECT id, fingerprint
	        FROM file
	        WHERE directory = ? AND name = ?`

	rows, err := db.connection.Query(sql, directory, name)
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
	var fingerprint string
	err = rows.Scan(&id, &fingerprint)
	if err != nil {
		return nil, err
	}

	return &entities.File{id, directory, name, fingerprint}, nil
}

func (db Database) FilesByDirectory(path string) ([]entities.File, error) {
    directory := filepath.Dir(path)
    name := filepath.Base(path)

    sql := `SELECT id, directory, name, fingerprint
            FROM file
            WHERE directory = ? AND name = ?
            OR directory like ?`

    rows, err := db.connection.Query(sql, directory, name, directory+"/%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]entities.File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var dir string
		var name string
		var fingerprint string
		err = rows.Scan(&fileId, &dir, &name, &fingerprint)
		if err != nil {
			return nil, err
		}

		files = append(files, entities.File{fileId, dir, name, fingerprint})
	}

	return files, nil
}

func (db Database) FilesByFingerprint(fingerprint string) ([]entities.File, error) {
	sql := `SELECT id, directory, name
	        FROM file
	        WHERE fingerprint = ?`

	rows, err := db.connection.Query(sql, fingerprint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]entities.File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		err = rows.Scan(&fileId, &directory, &name)
		if err != nil {
			return nil, err
		}

		files = append(files, entities.File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (db Database) DuplicateFiles() ([][]entities.File, error) {
	sql := `SELECT id, directory, name, fingerprint
            FROM file
            WHERE fingerprint IN (SELECT fingerprint
                                FROM file
                                GROUP BY fingerprint
                                HAVING count(1) > 1)
            ORDER BY fingerprint`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileSets := make([][]entities.File, 0, 10)
	var fileSet []entities.File
	var previousFingerprint string

	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		err = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if err != nil {
			return nil, err
		}

		if fingerprint != previousFingerprint {
			if fileSet != nil {
				fileSets = append(fileSets, fileSet)
			}
			fileSet = make([]entities.File, 0, 10)
			previousFingerprint = fingerprint
		}

		fileSet = append(fileSet, entities.File{fileId, directory, name, fingerprint})
	}

	// ensure last file set is added
	if len(fileSet) > 0 {
		fileSets = append(fileSets, fileSet)
	}

	return fileSets, nil
}

func (db Database) AddFile(path string, fingerprint string) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `INSERT INTO file (directory, name, fingerprint)
	        VALUES (?, ?, ?)`

	result, err := db.connection.Exec(sql, directory, name, fingerprint)
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

	return &entities.File{uint(id), directory, name, fingerprint}, nil
}

func (db Database) FileCountWithTags(tagNames []string) (uint, error) {
	sql := `SELECT count(1)
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
                HAVING count(1) = ` + strconv.Itoa(len(tagNames)) + `
            )`

	// convert string array to empty-interface array
	castTagNames := make([]interface{}, len(tagNames))
	for index, tagName := range tagNames {
		castTagNames[index] = tagName
	}

	rows, err := db.connection.Query(sql, castTagNames...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("Count query returned no rows.")
	}
	if rows.Err() != nil {
		return 0, err
	}

	var fileCount uint
	err = rows.Scan(&fileCount)
	if err != nil {
		return 0, err
	}

	return fileCount, nil
}

func (db Database) FilesWithTags(tagNames []string) ([]entities.File, error) {
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
                HAVING count(1) = ` + strconv.Itoa(len(tagNames)) + `
            )`

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

	files := make([]entities.File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fingerprint string
		err = rows.Scan(&fileId, &directory, &name, &fingerprint)
		if err != nil {
			return nil, err
		}

		files = append(files, entities.File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (db Database) UpdateFileFingerprint(fileId uint, fingerprint string) error {
	sql := `UPDATE file
	        SET fingerprint = ?
	        WHERE id = ?`

	_, err := db.connection.Exec(sql, fingerprint, int(fileId))
	if err != nil {
		return err
	}

	return nil
}

func (db Database) RemoveFile(fileId uint) error {
	sql := `DELETE FROM file
	        WHERE id = ?`

	result, err := db.connection.Exec(sql, fileId)
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

func (db Database) FileTagCount() (uint, error) {
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

func (db Database) FileTags() ([]entities.FileTag, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags := make([]entities.FileTag, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileTagId uint
		var fileId uint
		var tagId uint
		err = rows.Scan(&fileTagId, &fileId, &tagId)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, entities.FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
}

func (db Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*entities.FileTag, error) {
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

	return &entities.FileTag{fileTagId, fileId, tagId}, nil
}

func (db Database) FileTagsByTagId(tagId uint) ([]entities.FileTag, error) {
	sql := `SELECT id, file_id
	        FROM file_tag
	        WHERE tag_id = ?`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags := make([]entities.FileTag, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileTagId uint
		var fileId uint
		err = rows.Scan(&fileTagId, &fileId)
		if err != nil {
			return nil, err
		}

		fileTags = append(fileTags, entities.FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
}

func (db Database) AnyFileTagsForFile(fileId uint) (bool, error) {
	sql := `SELECT 1
            FROM file_tag
            WHERE file_id = ?
            LIMIT 1`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	if rows.Err() != nil {
		return false, err
	}

	return false, nil
}

func (db Database) AddFileTag(fileId uint, tagId uint) (*entities.FileTag, error) {
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

	return &entities.FileTag{uint(id), fileId, tagId}, nil
}

func (db Database) RemoveFileTag(fileId uint, tagId uint) error {
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

func (db Database) RemoveFileTagsByFileId(fileId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE file_id = ?`

	_, err := db.connection.Exec(sql, fileId)
	if err != nil {
		return err
	}

	return nil
}

func (db Database) RemoveFileTagsByTagId(tagId uint) error {
	sql := `DELETE FROM file_tag
	        WHERE tag_id = ?`

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

func (db Database) UpdateFileTags(oldTagId uint, newTagId uint) error {
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

func (db Database) CreateSchema() error {
	sql := `CREATE TABLE IF NOT EXISTS tag (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )`

	_, err := db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_tag_name
           ON tag(name)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file (
               id INTEGER PRIMARY KEY,
               directory TEXT NOT NULL,
               name TEXT NOT NULL,
               fingerprint TEXT NOT NULL,
               CONSTRAINT con_file_path UNIQUE (directory, name)
           )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_fingerprint
           ON file(fingerprint)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_path
           ON file(directory, name)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS file_tag (
               id INTEGER PRIMARY KEY,
               file_id INTEGER NOT NULL,
               tag_id INTEGER NOT NULL,
               FOREIGN KEY (file_id) REFERENCES file(id),
               FOREIGN KEY (tag_id) REFERENCES tag(id)
           )`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
           ON file_tag(file_id)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
           ON file_tag(tag_id)`

	_, err = db.connection.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (db Database) contains(list []string, str string) bool {
	for _, current := range list {
		if current == str {
			return true
		}
	}

	return false
}
