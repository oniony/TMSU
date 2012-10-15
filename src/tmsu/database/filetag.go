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
	"os"
	"path/filepath"
	"time"
	"tmsu/common"
)

type FileTag struct {
	Id     uint
	FileId uint
	TagId  uint
}

func (db Database) FileCountWithTags(tagIds []uint, explicit bool) (uint, error) {
	files, err := db.FilesWithTags(tagIds, []uint{}, explicit)
	if err != nil {
		return 0, err
	}

	return uint(len(files)), nil
}

func (db Database) FilesWithTag(tagId uint, explicit bool) (Files, error) {
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

	files, err := readFiles(rows, make(Files, 0, 10))
	if err != nil {
		return nil, err
	}

	if !explicit {
		for index := 0; index < len(files); index += 1 {
			file := files[index]

			additionalPaths, err := common.DirectoryEntries(file.Path())
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}

			if additionalPaths != nil {
				for _, additionalPath := range additionalPaths {
					directory := filepath.Dir(additionalPath)
					filename := filepath.Base(additionalPath)
					files = append(files, File{0, directory, filename, "", time.Time{}})
				}
			}
		}
	}

	return files, nil
}

func (db Database) FilesWithTags(includeTagIds, excludeTagIds []uint, explicit bool) (Files, error) {
	var fileByPath map[string]File = make(map[string]File, 10)

	if len(includeTagIds) > 0 {
		files, err := db.FilesWithTag(includeTagIds[0], explicit)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			fileByPath[file.Path()] = file
		}

		for _, includeTagId := range includeTagIds[1:] {
			files, err := db.FilesWithTag(includeTagId, explicit)
			if err != nil {
				return nil, err
			}

			for path, file := range fileByPath {
				if !contains(files, file) {
					delete(fileByPath, path)
				}
			}
		}
	} else {
		files, err := db.Files()
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			fileByPath[file.Path()] = file
		}
	}

	for _, excludeTagId := range excludeTagIds {
		files, err := db.FilesWithTag(excludeTagId, explicit)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			_, contains := fileByPath[file.Path()]
			if contains {
				delete(fileByPath, file.Path())
			}
		}
	}

	files := make(Files, len(fileByPath))
	for _, file := range fileByPath {
		files = append(files, file)
	}

	return files, nil
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

func (db Database) FileTags() (FileTags, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFileTags(rows, make(FileTags, 0, 10))
}

func (db Database) FileTagByFileIdAndTagId(fileId uint, tagId uint) (*FileTag, error) {
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

func (db Database) FileTagsByTagId(tagId uint) (FileTags, error) {
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

func (db Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
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

type FileTags []FileTag

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

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
}

func contains(files Files, searchFile File) bool {
	for _, file := range files {
		if file.Path() == searchFile.Path() {
			return true
		}
	}

	return false
}
