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
    "path/filepath"
    "tmsu/fingerprint"
)

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

func (db Database) Files() ([]*File, error) {
	sql := `SELECT id, directory, name, fingerprint
	        FROM file`

	rows, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]*File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fp string
		err = rows.Scan(&fileId, &directory, &name, &fp)
		if err != nil {
			return nil, err
		}

		files = append(files, &File{fileId, directory, name, fingerprint.Fingerprint(fp)})
	}

	return files, nil
}

func (db Database) File(id uint) (*File, error) {
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
	var fp string
	err = rows.Scan(&directory, &name, &fp)
	if err != nil {
		return nil, err
	}

	return &File{id, directory, name, fingerprint.Fingerprint(fp)}, nil
}

func (db Database) FileByPath(path string) (*File, error) {
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
	var fp string
	err = rows.Scan(&id, &fp)
	if err != nil {
		return nil, err
	}

	return &File{id, directory, name, fingerprint.Fingerprint(fp)}, nil
}

func (db Database) FilesByDirectory(path string) ([]*File, error) {
	sql := `SELECT id, directory, name, fingerprint
            FROM file
            WHERE directory = ? OR directory LIKE ?`

	rows, err := db.connection.Query(sql, path, filepath.Clean(path + "/%"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]*File, 0, 10)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var dir string
		var name string
		var fp string
		err = rows.Scan(&fileId, &dir, &name, &fp)
		if err != nil {
			return nil, err
		}
		files = append(files, &File{fileId, dir, name, fingerprint.Fingerprint(fp)})
	}

	return files, nil
}

func (db Database) FilesByFingerprint(fingerprint fingerprint.Fingerprint) ([]File, error) {
	sql := `SELECT id, directory, name
	        FROM file
	        WHERE fingerprint = ?`

	rows, err := db.connection.Query(sql, string(fingerprint))
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
		err = rows.Scan(&fileId, &directory, &name)
		if err != nil {
			return nil, err
		}

		files = append(files, File{fileId, directory, name, fingerprint})
	}

	return files, nil
}

func (db Database) DuplicateFiles() ([][]File, error) {
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

	fileSets := make([][]File, 0, 10)
	var fileSet []File
	var previousFingerprint fingerprint.Fingerprint

	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var fileId uint
		var directory string
		var name string
		var fp string
		err = rows.Scan(&fileId, &directory, &name, &fp)
		if err != nil {
			return nil, err
		}

        fingerprint := fingerprint.Fingerprint(fp)

		if fingerprint != previousFingerprint {
			if fileSet != nil {
				fileSets = append(fileSets, fileSet)
			}
			fileSet = make([]File, 0, 10)
			previousFingerprint = fingerprint
		}

		fileSet = append(fileSet, File{fileId, directory, name, fingerprint})
	}

	// ensure last file set is added
	if len(fileSet) > 0 {
		fileSets = append(fileSets, fileSet)
	}

	return fileSets, nil
}

func (db Database) AddFile(path string, fingerprint fingerprint.Fingerprint) (*File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `INSERT INTO file (directory, name, fingerprint)
	        VALUES (?, ?, ?)`

	result, err := db.connection.Exec(sql, directory, name, string(fingerprint))
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

	return &File{uint(id), directory, name, fingerprint}, nil
}

func (db Database) UpdateFileFingerprint(fileId uint, fingerprint fingerprint.Fingerprint) error {
	sql := `UPDATE file
	        SET fingerprint = ?
	        WHERE id = ?`

	_, err := db.connection.Exec(sql, string(fingerprint), int(fileId))
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

