/*
Copyright 2011-2014 Paul Ruane.

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
	"tmsu/common/fingerprint"
	"tmsu/entities"
	"tmsu/query"
)

// Retrieves the total number of tracked files.
func (db *Database) FileCount() (uint, error) {
	sql := `SELECT count(1)
			FROM file`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// The complete set of tracked files.
func (db *Database) Files() (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        ORDER BY directory || '/' || name`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves a specific file.
func (db *Database) File(id uint) (*entities.File, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE id = ?`

	rows, err := db.transaction.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves the file with the specified path.
func (db *Database) FileByPath(path string) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE directory = ? AND name = ?`

	rows, err := db.transaction.Query(sql, directory, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves all files that are under the specified directory.
func (db *Database) FilesByDirectory(path string) (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE directory = ? OR directory LIKE ?
            ORDER BY directory || '/' || name`

	rows, err := db.transaction.Query(sql, path, filepath.Clean(path+"/%"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the number of files with the specified fingerprint.
func (db *Database) FileCountByFingerprint(fingerprint fingerprint.Fingerprint) (uint, error) {
	sql := `SELECT count(id)
            FROM file
            WHERE fingerprint = ?`

	rows, err := db.transaction.Query(sql, string(fingerprint))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files with the specified fingerprint.
func (db *Database) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE fingerprint = ?
	        ORDER BY directory || '/' || name`

	rows, err := db.transaction.Query(sql, string(fingerprint))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 1))
}

// Retrieves the set of untagged files.
func (db *Database) UntaggedFiles() (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE id NOT IN (SELECT distinct(file_id)
                             FROM file_tag)`

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the count of files matching the specified query.
func (db *Database) QueryFileCount(expression query.Expression) (uint, error) {
	sql := buildCountQuery(expression)

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files matching the specified query.
func (db *Database) QueryFiles(expression query.Expression) (entities.Files, error) {
	sql := buildQuery(expression)
	rows, err := db.transaction.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the sets of duplicate files within the database.
func (db *Database) DuplicateFiles() ([]entities.Files, error) {
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

	rows, err := db.transaction.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fileSets := make([]entities.Files, 0, 10)
	var fileSet entities.Files
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
			fileSet = make(entities.Files, 0, 10)
			previousFingerprint = fingerprint
		}

		fileSet = append(fileSet, &entities.File{fileId, directory, name, fingerprint, modTime, size, isDir})
	}

	// ensure last file set is added
	if len(fileSet) > 0 {
		fileSets = append(fileSets, fileSet)
	}

	return fileSets, nil
}

// Adds a file to the database.
func (db *Database) InsertFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `INSERT INTO file (directory, name, fingerprint, mod_time, size, is_dir)
	        VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.transaction.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir)
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

	return &entities.File{uint(id), directory, name, fingerprint, modTime, size, isDir}, nil
}

// Updates a file in the database.
func (db *Database) UpdateFile(fileId uint, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `UPDATE file
	        SET directory = ?, name = ?, fingerprint = ?, mod_time = ?, size = ?, is_dir = ?
	        WHERE id = ?`

	result, err := db.transaction.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir, int(fileId))
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

	return &entities.File{uint(fileId), directory, name, fingerprint, modTime, size, isDir}, nil
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

	result, err := db.transaction.Exec(sql, fileId)
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

// Deletes all untagged files.
func (db *Database) DeleteUntaggedFiles() error {
	sql := `DELETE FROM file
            WHERE id NOT IN (SELECT distinct(file_id)
                             FROM file_tag)`

	_, err := db.transaction.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// unexported

func readFile(rows *sql.Rows) (*entities.File, error) {
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

	return &entities.File{fileId, directory, name, fingerprint.Fingerprint(fp), modTime, size, isDir}, nil
}

func readFiles(rows *sql.Rows, files entities.Files) (entities.Files, error) {
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

func buildCountQuery(expression query.Expression) string {
	builder := NewBuilder()

	builder.AppendSql("SELECT count(id) FROM file WHERE 1 == 1 AND\n")
	buildQueryBranch(expression, builder)

	return builder.Sql
}

func buildQuery(expression query.Expression) string {
	builder := NewBuilder()

	builder.AppendSql("SELECT id, directory, name, fingerprint, mod_time, size, is_dir FROM file WHERE 1==1 AND\n")
	buildQueryBranch(expression, builder)
	builder.AppendSql("ORDER BY directory || '/' || name")

	return builder.Sql
}

func buildQueryBranch(expression query.Expression, builder *SqlBuilder) {
	switch exp := expression.(type) {
	case query.TagExpression:
		builder.AppendSql(`id IN (SELECT file_id
FROM file_tag
WHERE tag_id = (SELECT id
                FROM tag
                WHERE name = '` + exp.Name + `'))
`)
	case query.ComparisonExpression:
		var value, valueExpression string
		_, err := strconv.ParseFloat(exp.Value.Name, 64)
		if err == nil {
			value = exp.Value.Name
			valueExpression = "CAST(name AS float)"
		} else {
			value = "'" + exp.Value.Name + "'"
			valueExpression = "name"
		}

		builder.AppendSql(`id IN (SELECT file_id
FROM file_tag
WHERE tag_id = (SELECT id
                FROM tag
                WHERE name = '` + exp.Tag.Name + `')
AND value_id IN (SELECT id
                 FROM value
                 WHERE ` + valueExpression + ` ` + exp.Operator + ` ` + value + `))`)
	case query.NotExpression:
		builder.AppendSql("\nNOT\n")
		buildQueryBranch(exp.Operand, builder)
	case query.AndExpression:
		buildQueryBranch(exp.LeftOperand, builder)
		builder.AppendSql("\nAND\n")
		buildQueryBranch(exp.RightOperand, builder)
	case query.OrExpression:
		builder.AppendSql("(\n")
		buildQueryBranch(exp.LeftOperand, builder)
		builder.AppendSql("\nOR\n")
		buildQueryBranch(exp.RightOperand, builder)
		builder.AppendSql(")\n")
	case query.EmptyExpression:
		builder.AppendSql("1 == 1\n")
	default:
		panic("Unsupported expression type.")
	}
}
