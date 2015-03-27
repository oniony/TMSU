// Copyright 2011-2015 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"time"
	"tmsu/common/fingerprint"
	"tmsu/entities"
	"tmsu/query"
)

// Retrieves the total number of tracked files.
func FileCount(tx *Tx) (uint, error) {
	sql := `SELECT count(1)
			FROM file`

	rows, err := tx.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// The complete set of tracked files.
func Files(tx *Tx, sort string) (entities.Files, error) {
	builder := NewBuilder()
	builder.AppendSql(`SELECT id, directory, name, fingerprint, mod_time, size, is_dir
                       FROM file `)

	buildSort(sort, builder)

	rows, err := tx.Query(builder.Sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves a specific file.
func File(tx *Tx, id entities.FileId) (*entities.File, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE id = ?`

	rows, err := tx.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves the file with the specified path.
func FileByPath(tx *Tx, path string) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE directory = ? AND name = ?`

	rows, err := tx.Query(sql, directory, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFile(rows)
}

// Retrieves all files that are under the specified directory.
func FilesByDirectory(tx *Tx, path string) (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE directory = ? OR directory LIKE ?
            ORDER BY directory || '/' || name`

	path = filepath.Clean(path)

	rows, err := tx.Query(sql, path, filepath.Join(path, "%"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the number of files with the specified fingerprint.
func FileCountByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (uint, error) {
	sql := `SELECT count(id)
            FROM file
            WHERE fingerprint = ?`

	rows, err := tx.Query(sql, string(fingerprint))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files with the specified fingerprint.
func FilesByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
	        FROM file
	        WHERE fingerprint = ?
	        ORDER BY directory || '/' || name`

	rows, err := tx.Query(sql, string(fingerprint))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 1))
}

// Retrieves the set of untagged files.
func UntaggedFiles(tx *Tx) (entities.Files, error) {
	sql := `SELECT id, directory, name, fingerprint, mod_time, size, is_dir
            FROM file
            WHERE id NOT IN (SELECT distinct(file_id)
                             FROM file_tag)`

	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the count of files matching the specified query and matching the specified path.
func QueryFileCount(tx *Tx, expression query.Expression, path string) (uint, error) {
	builder := buildCountQuery(expression, path)

	rows, err := tx.Query(builder.Sql, builder.Params...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readCount(rows)
}

// Retrieves the set of files matching the specified query and matching the specified path.
func QueryFiles(tx *Tx, expression query.Expression, path, sort string) (entities.Files, error) {
	builder := buildQuery(expression, path, sort)
	rows, err := tx.Query(builder.Sql, builder.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readFiles(rows, make(entities.Files, 0, 10))
}

// Retrieves the sets of duplicate files within the database.
func DuplicateFiles(tx *Tx) ([]entities.Files, error) {
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

	rows, err := tx.Query(sql)
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

		var fileId entities.FileId
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
func InsertFile(tx *Tx, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `INSERT INTO file (directory, name, fingerprint, mod_time, size, is_dir)
	        VALUES (?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir)
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
		panic("expected exactly one row to be affected.")
	}

	return &entities.File{entities.FileId(id), directory, name, fingerprint, modTime, size, isDir}, nil
}

// Updates a file in the database.
func UpdateFile(tx *Tx, fileId entities.FileId, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	sql := `UPDATE file
	        SET directory = ?, name = ?, fingerprint = ?, mod_time = ?, size = ?, is_dir = ?
	        WHERE id = ?`

	result, err := tx.Exec(sql, directory, name, string(fingerprint), modTime, size, isDir, int(fileId))
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		panic("expected exactly one row to be affected.")
	}

	return &entities.File{entities.FileId(fileId), directory, name, fingerprint, modTime, size, isDir}, nil
}

// Removes a file from the database.
func DeleteFile(tx *Tx, fileId entities.FileId) error {
	sql := `DELETE FROM file
	        WHERE id = ?`

	result, err := tx.Exec(sql, fileId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return NoSuchFileError{fileId}
	}
	if rowsAffected > 1 {
		panic("expected only one row to be affected.")
	}

	return nil
}

// Deletes the specified files if they are untagged
func DeleteUntaggedFiles(tx *Tx, fileIds entities.FileIds) error {
	for _, fileId := range fileIds {
		sql := `DELETE FROM file
                WHERE id = ?1
                AND (SELECT count(1)
                     FROM file_tag
                     WHERE file_id = ?1) == 0`

		_, err := tx.Exec(sql, fileId)
		if err != nil {
			return err
		}
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

	var fileId entities.FileId
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

func buildCountQuery(expression query.Expression, path string) *SqlBuilder {
	builder := NewBuilder()

	builder.AppendSql("SELECT count(id) FROM file WHERE 1 == 1 AND\n")
	buildQueryBranch(expression, builder)
	buildPathClause(path, builder)

	return builder
}

func buildQuery(expression query.Expression, path, sort string) *SqlBuilder {
	builder := NewBuilder()

	builder.AppendSql("SELECT id, directory, name, fingerprint, mod_time, size, is_dir FROM file WHERE 1==1 AND\n")
	buildQueryBranch(expression, builder)
	buildPathClause(path, builder)
	buildSort(sort, builder)

	return builder
}

func buildQueryBranch(expression query.Expression, builder *SqlBuilder) {
	switch exp := expression.(type) {
	case query.TagExpression:
		builder.AppendSql(`id IN (SELECT file_id FROM file_tag WHERE tag_id = (SELECT id FROM tag WHERE name = `)
		builder.AppendParam(exp.Name)
		builder.AppendSql(`))`)
	case query.ComparisonExpression:
		var valueExpression string
		_, err := strconv.ParseFloat(exp.Value.Name, 64)
		if err == nil {
			valueExpression = "CAST(name AS float)"
		} else {
			valueExpression = "name"
		}

		builder.AppendSql(`id IN (SELECT file_id FROM file_tag WHERE tag_id = (SELECT id FROM tag WHERE name = `)
		builder.AppendParam(exp.Tag.Name)
		builder.AppendSql(`) AND value_id IN (SELECT id FROM value WHERE ` + valueExpression + ` ` + exp.Operator + ` `)
		builder.AppendParam(exp.Value.Name)
		builder.AppendSql(`))`)
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

func buildPathClause(path string, builder *SqlBuilder) {
	path = filepath.Clean(path)

	dir, name := filepath.Split(path)
	dir = filepath.Clean(dir)

	if path != "" {
		builder.AppendSql("AND (directory = ")
		builder.AppendParam(path)
		builder.AppendSql(" OR directory LIKE ")
		builder.AppendParam(filepath.Join(path, "%"))
		builder.AppendSql(" OR (directory = ")
		builder.AppendParam(dir)
		builder.AppendSql(" AND name = ")
		builder.AppendParam(name)
		builder.AppendSql("))\n")

		//builder.AppendSql("AND (directory = '" + path + "' OR directory LIKE '" + filepath.Join(path, "%") + "'x OR (directory = '" + dir + "' AND name = '" + name + "'))\n")
	}
}

func buildSort(sort string, builder *SqlBuilder) {
	switch sort {
	case "none":
		// do nowt
	case "id":
		builder.AppendSql("ORDER BY id")
	case "name":
		builder.AppendSql("ORDER BY directory || '/' || name")
	case "time":
		builder.AppendSql("ORDER BY mod_time, directory || '/' || name")
	case "size":
		builder.AppendSql("ORDER BY size, directory || '/' || name")
	}
}
