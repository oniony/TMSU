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
	"errors"
	"strconv"
	"strings"
	"tmsu/fingerprint"
)

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

func (db Database) FilesWithTags(includeTagNames, excludeTagNames []string) ([]File, error) {
	castTagNames := make([]interface{}, len(includeTagNames)+len(excludeTagNames))
	for index, tagName := range includeTagNames {
		castTagNames[index] = tagName
	}
	for index, tagName := range excludeTagNames {
		castTagNames[index+len(includeTagNames)] = tagName
	}

	sql := `SELECT id, directory, name, fingerprint, mod_time
			FROM file
			WHERE 1 = 1`

	if len(includeTagNames) > 0 {
		sql += ` AND id IN (
					SELECT file_id
					FROM file_tag
					WHERE tag_id IN (
						SELECT id
						FROM tag
						WHERE name IN (` + strings.Repeat("?,", len(includeTagNames)-1) + `?)
					)
					GROUP BY file_id
					HAVING count(1) = ` + strconv.Itoa(len(includeTagNames)) + `
				)`
	}

	if len(excludeTagNames) > 0 {
		sql += ` AND id NOT IN (
					SELECT file_id
					FROM file_tag
					WHERE tag_id IN (
						SELECT id
						FROM tag
						WHERE name IN (` + strings.Repeat("?,", len(excludeTagNames)-1) + `?)
					)
				)`
	}

	rows, err := db.connection.Query(sql, castTagNames...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fp string
		var modTime string
		err = rows.Scan(&fileId, &directory, &name, &fp, &modTime)
		if err != nil {
			return nil, err
		}

		files = append(files, File{fileId, directory, name, fingerprint.Fingerprint(fp), parseTimestamp(modTime)})
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

func (db Database) FileTags() ([]FileTag, error) {
	sql := `SELECT id, file_id, tag_id
	        FROM file_tag`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
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

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
	}

	return fileTags, nil
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

func (db Database) FileTagsByTagId(tagId uint) ([]FileTag, error) {
	sql := `SELECT id, file_id
	        FROM file_tag
	        WHERE tag_id = ?`

	rows, err := db.connection.Query(sql, tagId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileTags := make([]FileTag, 0, 10)
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

		fileTags = append(fileTags, FileTag{fileTagId, fileId, tagId})
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
