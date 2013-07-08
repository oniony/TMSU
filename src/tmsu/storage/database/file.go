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
	"path/filepath"
	"strconv"
	"time"
	"tmsu/fingerprint"
)

// A tracked file.
type File struct {
	Id          uint
	Directory   string
	Name        string
	Fingerprint fingerprint.Fingerprint
	ModTime     time.Time
	Size        int64
	IsDir       bool
}

type Files []*File

func (files Files) Where(predicate func(*File) bool) Files {
	result := make(Files, 0, 10)

	for _, file := range files {
		if predicate(file) {
			result = append(result, file)
		}
	}

	return result
}

// Retrieves the file's path.
func (file File) Path() string {
	return filepath.Join(file.Directory, file.Name)
}

// Retrieves the total number of tracked files.
func (db *Database) FileCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// The complete set of tracked files.
func (db *Database) Files() (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        ORDER BY directory || '/' || name`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

// Retrieves a specific file.
func (db *Database) File(id uint) (*File, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE id = ?`

	rows, err := db.connection.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves the file with the specified path.
func (db *Database) FileByPath(path string) (*File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE directory = ? AND name = ?`

	rows, err := db.connection.Query(sql, directory, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves all files that are under the specified directory.
func (db *Database) FilesByDirectory(path string) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE directory = ? OR directory LIKE ?
            ORDER BY directory || '/' || name`

	rows, err := db.connection.Query(sql, path, filepath.Clean(path+"/%"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 1))
}

// Retrieves the number of files with the specified fingerprint.
func (db *Database) FileCountByFingerprint(fingerprint fingerprint.Fingerprint) (uint, error) {
	sql := `SELECT count(id)
            FROM file
            WHERE fingerprint = ?`

	rows, err := db.connection.Query(sql, string(fingerprint))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files with the specified fingerprint.
func (db *Database) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE fingerprint = ?
	        ORDER BY directory || '/' || name`

	rows, err := db.connection.Query(sql, string(fingerprint))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 1))
}

// Retrieves the count of files with the specified tag.
func (db *Database) FileCountWithTag(tagId uint) (uint, error) {
	sql := `SELECT count(1)
            FROM file_tag
            WHERE tag_id == ?`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files with the specified tag.
func (db *Database) FilesWithTag(tagId uint) (Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id = ?1
		    )
            ORDER BY directory || '/' || name`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

// Retrieves the count of files with the specified tags.
func (db *Database) FileCountWithTags(tagIds []uint) (uint, error) {
	tagCount := len(tagIds)

	sql := `SELECT count(1)
	        FROM (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (?1`

	for idx := 2; idx <= tagCount; idx += 1 {
		sql += ", ?" + strconv.Itoa(idx)
	}

	sql += `    )
                GROUP BY file_id
                HAVING count(tag_id) == ?` + strconv.Itoa(tagCount+1) + `
            )`

	params := make([]interface{}, tagCount+1)
	for index, tagId := range tagIds {
		params[index] = interface{}(tagId)
	}
	params[tagCount] = tagCount

	rows, err := db.connection.Query(sql, params...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files with the specified tags.
func (db *Database) FilesWithTags(includeTagIds []uint, excludeTagIds []uint) (Files, error) {
	includeTagCount := len(includeTagIds)
	excludeTagCount := len(excludeTagIds)

	paramIndex := 1

	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE 1==1`

	if includeTagCount > 0 {
		sql += `
            AND id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (`

		for idx := 0; idx < includeTagCount; idx++ {
			if idx != 0 {
				sql += ","
			}

			sql += "?" + strconv.Itoa(paramIndex)
			paramIndex++
		}

		sql += `)
                GROUP BY file_id
                HAVING count(tag_id) == ?` + strconv.Itoa(paramIndex) + `
            )`

		paramIndex++
	}

	if excludeTagCount > 0 {
		sql += `
            AND id NOT IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (`

		for idx := 0; idx < excludeTagCount; idx++ {
			if idx != 0 {
				sql += ","
			}

			sql += "?" + strconv.Itoa(paramIndex)
			paramIndex++
		}

		sql += `)
            )`
	}

	sql += `ORDER BY directory || '/' || name`

	params := make([]interface{}, paramIndex-1)
	paramIndex = 0
	for _, includeTagId := range includeTagIds {
		params[paramIndex] = interface{}(includeTagId)
		paramIndex++
	}

	if includeTagCount > 0 {
		params[paramIndex] = interface{}(includeTagCount)
		paramIndex++
	}

	for _, excludeTagId := range excludeTagIds {
		params[paramIndex] = interface{}(excludeTagId)
		paramIndex++
	}

	rows, err := db.connection.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(Files, 0, 10))
}

// Retrieves the sets of duplicate files within the database.
func (db *Database) DuplicateFiles() ([]Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE fingerprint IN (
                SELECT fingerprint
                FROM file
                WHERE fingerprint != ''
                GROUP BY fingerprint
                HAVING count(1) > 1
            )
            ORDER BY fingerprint, directory || '/' || name`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileSets := make([]Files, 0, 10)
	var fileSet Files
	var previousFingerprint fingerprint.Fingerprint

	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory, name, fp string
		var modTime time.Time
		var size int64
		var isDir bool
		err = rows.Scan(&fileId, &directory, &name, &fp, &modTime, &size, &isDir)
		if err != nil {
			return nil, err
		}

		fingerprint := fingerprint.Fingerprint(fp)

		if fingerprint != previousFingerprint {
			if fileSet != nil {
				fileSets = append(fileSets, fileSet)
			}
			fileSet = make(Files, 0, 10)
			previousFingerprint = fingerprint
		}

		fileSet = append(fileSet, &File{fileId, directory, name, fingerprint, modTime, size, isDir})
	}

	// ensure last file set is added
	if len(fileSet) > 0 {
		fileSets = append(fileSets, fileSet)
	}

	return fileSets, nil
}

// Adds a file to the database.
func (db *Database) InsertFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `INSERT INTO file (directory, name, fingerprint, mod_time, size, is_dir)
	        VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.connection.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir)
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
		return nil, errors.New("expected exactly one row to be affected.")
	}

	return &File{uint(id), directory, name, fingerprint, modTime, size, isDir}, nil
}

// Updates a file in the database.
func (db *Database) UpdateFile(fileId uint, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `UPDATE file
	        SET directory = ?, name = ?, fingerprint = ?, mod_time = ?, size = ?, is_dir = ?
	        WHERE id = ?`

	result, err := db.connection.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir, int(fileId))
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		return nil, errors.New("expected exactly one row to be affected.")
	}

	return &File{uint(fileId), directory, name, fingerprint, modTime, size, isDir}, nil
}

// Removes a file from the database.
func (db *Database) DeleteFile(fileId uint) error {
	file, err := db.File(fileId)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("no such file '" + strconv.Itoa(int(fileId)) + "'.")
	}

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
	if rowsAffected > 1 {
		return errors.New("expected only one row to be affected.")
	}

	return nil
}

//

func readFile(rows *sql.Rows) (*File, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var fileId uint
	var directory, name, fp string
	var modTime time.Time
	var size int64
	var isDir bool
	err := rows.Scan(&fileId, &directory, &name, &fp, &modTime, &size, &isDir)
	if err != nil {
		return nil, err
	}

	return &File{fileId, directory, name, fingerprint.Fingerprint(fp), modTime, size, isDir}, nil
}

func readFiles(rows *sql.Rows, files Files) (Files, error) {
	for {
		file, err := readFile(rows)
		if err != nil {
			return nil, err
		}
		if file == nil {
			break
		}

		files = append(files, file)
	}

	return files, nil
}
