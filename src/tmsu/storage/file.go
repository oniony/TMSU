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

package storage

import (
	"fmt"
	"path/filepath"
	"time"
	"tmsu/common/fingerprint"
	_path "tmsu/common/path"
	"tmsu/entities"
	"tmsu/query"
	"tmsu/storage/database"
)

// Retrieves the total number of tracked files.
func (storage *Storage) FileCount(tx *Tx) (uint, error) {
	return database.FileCount(tx.tx)
}

// The complete set of tracked files.
func (storage *Storage) Files(tx *Tx, sort string) (entities.Files, error) {
	files, err := database.Files(tx.tx, sort)
	storage.absPaths(files)

	return files, err
}

// Retrieves a specific file.
func (storage *Storage) File(tx *Tx, id entities.FileId) (*entities.File, error) {
	file, err := database.File(tx.tx, id)
	storage.absPath(file)

	return file, err
}

// Retrieves the file with the specified path.
func (storage *Storage) FileByPath(tx *Tx, path string) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := database.FileByPath(tx.tx, relPath)
	storage.absPath(file)

	return file, err
}

// Retrieves all files that are under the specified directory.
func (storage *Storage) FilesByDirectory(tx *Tx, path string) (entities.Files, error) {
	relPath := storage.relPath(path)
	files, err := database.FilesByDirectory(tx.tx, relPath)
	storage.absPaths(files)

	return files, err
}

// Retrieves all file that are under the specified directories.
func (storage *Storage) FilesByDirectories(tx *Tx, paths []string) (entities.Files, error) {
	files := make(entities.Files, 0, 100)

	for _, path := range paths {
		relPath := storage.relPath(path)
		pathFiles, err := database.FilesByDirectory(tx.tx, relPath)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve files for directory: %v", path, err)
		}

		files = append(files, pathFiles...)
	}

	storage.absPaths(files)

	return files, nil
}

// Retrieves the number of files with the specified fingerprint.
func (storage *Storage) FileCountByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (uint, error) {
	return database.FileCountByFingerprint(tx.tx, fingerprint)
}

// Retrieves the set of files with the specified fingerprint.
func (storage *Storage) FilesByFingerprint(tx *Tx, fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	files, err := database.FilesByFingerprint(tx.tx, fingerprint)
	storage.absPaths(files)
	return files, err
}

// Retrieves the set of untagged files.
func (storage *Storage) UntaggedFiles(tx *Tx) (entities.Files, error) {
	files, err := database.UntaggedFiles(tx.tx)
	storage.absPaths(files)
	return files, err
}

// Retrieves the count of files with the specified tags and matching the specified path.
func (storage *Storage) FileCountWithTags(tx *Tx, tagNames []string, path string, explicitOnly bool) (uint, error) {
	expression := query.HasAll(tagNames)

	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(tx, expression)
		if err != nil {
			return 0, err
		}
	}

	relPath := storage.relPath(path)
	return database.QueryFileCount(tx.tx, expression, relPath)
}

// Retrieves the count of files that match the specified query and matching the specified path.
func (storage *Storage) QueryFileCount(tx *Tx, expression query.Expression, path string, explicitOnly bool) (uint, error) {
	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(tx, expression)
		if err != nil {
			return 0, err
		}
	}

	relPath := storage.relPath(path)
	return database.QueryFileCount(tx.tx, expression, relPath)
}

// Retrieves the set of files that match the specified query.
func (storage *Storage) QueryFiles(tx *Tx, expression query.Expression, path string, explicitOnly bool, sort string) (entities.Files, error) {
	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(tx, expression)
		if err != nil {
			return nil, err
		}
	}

	relPath := storage.relPath(path)
	files, err := database.QueryFiles(tx.tx, expression, relPath, sort)
	storage.absPaths(files)
	return files, err
}

// Retrieves the sets of duplicate files within the database.
func (storage *Storage) DuplicateFiles(tx *Tx) ([]entities.Files, error) {
	fileSets, err := database.DuplicateFiles(tx.tx)

	for _, fileSet := range fileSets {
		storage.absPaths(fileSet)
	}

	return fileSets, err
}

// Adds a file to the database.
func (storage *Storage) AddFile(tx *Tx, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := database.InsertFile(tx.tx, relPath, fingerprint, modTime, size, isDir)
	storage.absPath(file)

	return file, err
}

// Updates a file in the database.
func (storage *Storage) UpdateFile(tx *Tx, fileId entities.FileId, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := database.UpdateFile(tx.tx, fileId, relPath, fingerprint, modTime, size, isDir)
	storage.absPath(file)

	return file, err
}

// Deletes a file from the database.
func (storage *Storage) DeleteFile(tx *Tx, fileId entities.FileId) error {
	return database.DeleteFile(tx.tx, fileId)
}

// Deletes a file if it is untagged
func (storage *Storage) DeleteFileIfUntagged(tx *Tx, fileId entities.FileId) error {
	count, err := storage.FileTagCountByFileId(tx, fileId, true)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.DeleteFile(tx, fileId); err != nil {
			return err
		}
	}

	return nil
}

// Deletes the specified files if they are untagged
func (storage *Storage) DeleteUntaggedFiles(tx *Tx, fileIds entities.FileIds) error {
	return database.DeleteUntaggedFiles(tx.tx, fileIds)
}

// unexported

func (storage *Storage) relPath(path string) string {
	if path == "" {
		return "" // don't alter empty paths
	}

	return _path.RelTo(path, storage.RootPath)
}

func (storage *Storage) absPaths(files entities.Files) {
	for _, file := range files {
		storage.absPath(file)
	}
}

func (storage *Storage) absPath(file *entities.File) {
	if file == nil || file.Directory == "" || file.Directory[0] == filepath.Separator {
		return
	}

	file.Directory = filepath.Join(storage.RootPath, file.Directory)
}

func (storage *Storage) addImpliedTags(tx *Tx, expression query.Expression) (query.Expression, error) {
	implications, err := storage.Implications(tx)
	if err != nil {
		fmt.Errorf("could not retrieve tag implications: %v", err)
	}

	impliersByTag := make(map[string][]string, len(implications))
	for _, implication := range implications {
		impliers, ok := impliersByTag[implication.ImpliedTag.Name]
		if !ok {
			impliers = make([]string, 0, 1)
		}

		impliersByTag[implication.ImpliedTag.Name] = append(impliers, implication.ImplyingTag.Name)
	}

	return storage.addImpliedTagsRecursive(expression, impliersByTag), nil
}

func (storage *Storage) addImpliedTagsRecursive(expression query.Expression, impliersByTag map[string][]string) query.Expression {
	switch typedExpression := expression.(type) {
	case query.OrExpression:
		typedExpression.LeftOperand = storage.addImpliedTagsRecursive(typedExpression.LeftOperand, impliersByTag)
		typedExpression.RightOperand = storage.addImpliedTagsRecursive(typedExpression.RightOperand, impliersByTag)
		return typedExpression
	case query.AndExpression:
		typedExpression.LeftOperand = storage.addImpliedTagsRecursive(typedExpression.LeftOperand, impliersByTag)
		typedExpression.RightOperand = storage.addImpliedTagsRecursive(typedExpression.RightOperand, impliersByTag)
		return typedExpression
	case query.NotExpression:
		typedExpression.Operand = storage.addImpliedTagsRecursive(typedExpression.Operand, impliersByTag)
		return typedExpression
	case query.TagExpression:
		return applyImplicationsForTag(typedExpression, impliersByTag)
	case query.ValueExpression, query.EmptyExpression, query.ComparisonExpression:
		return expression
	default:
		panic(fmt.Sprintf("unsupported expression type '%T'.", typedExpression))
	}
}

func applyImplicationsForTag(tagExpression query.TagExpression, impliersByTag map[string][]string) query.Expression {
	implyingTags, ok := impliersByTag[tagExpression.Name]
	if !ok {
		return tagExpression
	}

	var expression query.Expression = tagExpression

	for index := 0; index < len(implyingTags); index++ {
		implyingTag := implyingTags[index]

		expression = query.OrExpression{expression, query.TagExpression{implyingTag}}

		for _, furtherImplyingTag := range impliersByTag[implyingTag] {
			if furtherImplyingTag != tagExpression.Name && !containsTagName(implyingTags, furtherImplyingTag) {
				implyingTags = append(implyingTags, furtherImplyingTag)
			}
		}
	}

	return expression
}

func containsTagName(tagNames []string, tagName string) bool {
	for _, tn := range tagNames {
		if tn == tagName {
			return true
		}
	}

	return false
}
