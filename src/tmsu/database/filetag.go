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

func (db Database) FileCountWithTags(tagIds []uint) (uint, error) {
	//TODO optimize
	files, err := db.FilesWithTags(tagIds, []uint{})
	if err != nil {
		return 0, err
	}

	return uint(len(files)), nil
}

func (db Database) FilesWithTag(tagId uint) (Files, error) {
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

	explicitlyTaggedFiles, err := readFiles(rows, make(Files, 0, 10))
	if err != nil {
		return nil, err
	}

	files := make(Files, len(explicitlyTaggedFiles))
	for index, file := range explicitlyTaggedFiles {
		files[index] = file
	}

	for _, explicitlyTaggedFile := range explicitlyTaggedFiles {
		additionalFiles, err := db.FilesByDirectory(explicitlyTaggedFile.Path())
		if err != nil {
			return nil, err
		}

		for _, additionalFile := range additionalFiles {
			files = append(files, additionalFile)
		}
	}

	return files, nil
}

func (db Database) FilesWithTags(includeTagIds, excludeTagIds []uint) (Files, error) {
	var files Files

	if len(includeTagIds) > 0 {
		var err error
		files, err = db.FilesWithTag(includeTagIds[0])
		if err != nil {
			return nil, err
		}

		for _, tagId := range includeTagIds[1:] {
			filesWithTag, err := db.FilesWithTag(tagId)
			if err != nil {
				return nil, err
			}

			for index, file := range files {
				if !contains(filesWithTag, file) {
					files[index] = nil
				}
			}
		}
	}

	if len(excludeTagIds) > 0 {
		//TODO
	}

	resultFiles := make(Files, 0, len(files))
	for _, file := range files {
		if file != nil {
			resultFiles = append(resultFiles, file)
		}
	}

	return resultFiles, nil
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

func (db Database) TagsByFileId(fileId uint) (Tags, error) {
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

func (db Database) FileTagsByFileId(fileId uint, explicitOnly bool) (FileTags, error) {
	sql := `SELECT id, file_id, tag_id
            FROM file_tag
            WHERE file_id = ?`

	rows, err := db.connection.Query(sql, fileId)
	if err != nil {
		return nil, err
	}
	fileTags, err := readFileTags(rows, make(FileTags, 0, 10))
	rows.Close()

	if !explicitOnly {
		file, err := db.File(fileId)
		if err != nil {
			return nil, err
		}

		parentFile, err := db.FileByPath(file.Directory)
		if err != nil {
			return nil, err
		}

		if parentFile != nil {
			additionalFileTags, err := db.FileTagsByFileId(parentFile.Id, explicitOnly)
			if err != nil {
				return nil, err
			}

			for _, additionalFileTag := range additionalFileTags {
				fileTags = append(fileTags, additionalFileTag)
			}
		}
	}

	return fileTags, nil
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

type FileTags []*FileTag

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
