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
)

// Retrieves the total number of tracked files.
func (storage *Storage) FileCount() (uint, error) {
	return storage.Db.FileCount()
}

// The complete set of tracked files.
func (storage *Storage) Files() (entities.Files, error) {
	files, err := storage.Db.Files()
	storage.absPaths(files)

	return files, err
}

// Retrieves a specific file.
func (storage *Storage) File(id entities.FileId) (*entities.File, error) {
	file, err := storage.Db.File(id)
	storage.absPath(file)

	return file, err
}

// Retrieves the file with the specified path.
func (storage *Storage) FileByPath(path string) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := storage.Db.FileByPath(relPath)
	storage.absPath(file)

	return file, err
}

// Retrieves all files that are under the specified directory.
func (storage *Storage) FilesByDirectory(path string) (entities.Files, error) {
	relPath := storage.relPath(path)
	files, err := storage.Db.FilesByDirectory(relPath)
	storage.absPaths(files)

	return files, err
}

// Retrieves all file that are under the specified directories.
func (storage *Storage) FilesByDirectories(paths []string) (entities.Files, error) {
	files := make(entities.Files, 0, 100)

	for _, path := range paths {
		relPath := storage.relPath(path)
		pathFiles, err := storage.Db.FilesByDirectory(relPath)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve files for directory: %v", path, err)
		}

		files = append(files, pathFiles...)
	}

	storage.absPaths(files)

	return files, nil
}

// Retrieves the number of files with the specified fingerprint.
func (storage *Storage) FileCountByFingerprint(fingerprint fingerprint.Fingerprint) (uint, error) {
	return storage.Db.FileCountByFingerprint(fingerprint)
}

// Retrieves the set of files with the specified fingerprint.
func (storage *Storage) FilesByFingerprint(fingerprint fingerprint.Fingerprint) (entities.Files, error) {
	files, err := storage.Db.FilesByFingerprint(fingerprint)
	storage.absPaths(files)
	return files, err
}

// Retrieves the set of untagged files.
func (storage *Storage) UntaggedFiles() (entities.Files, error) {
	files, err := storage.Db.UntaggedFiles()
	storage.absPaths(files)
	return files, err
}

// Retrieves the count of files with the specified tags and matching the specified path.
func (storage *Storage) FileCountWithTags(tagNames []string, path string, explicitOnly bool) (uint, error) {
	expression := query.HasAll(tagNames)

	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(expression)
		if err != nil {
			return 0, err
		}
	}

	relPath := storage.relPath(path)
	return storage.Db.QueryFileCount(expression, relPath)
}

// Retrieves the set of files with the specified tags and matching the specified path.
func (storage *Storage) FilesWithTags(tagNames []string, path string, explicitOnly bool) (entities.Files, error) {
	expression := query.HasAll(tagNames)

	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(expression)
		if err != nil {
			return nil, err
		}
	}

	relPath := storage.relPath(path)
	files, err := storage.Db.QueryFiles(expression, relPath)
	storage.absPaths(files)
	return files, err
}

// Retrieves the count of files that match the specified query and matching the specified path.
func (storage *Storage) QueryFileCount(expression query.Expression, path string, explicitOnly bool) (uint, error) {
	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(expression)
		if err != nil {
			return 0, err
		}
	}

	relPath := storage.relPath(path)
	return storage.Db.QueryFileCount(expression, relPath)
}

// Retrieves the set of files that match the specified query.
func (storage *Storage) QueryFiles(expression query.Expression, path string, explicitOnly bool) (entities.Files, error) {
	if !explicitOnly {
		var err error
		expression, err = storage.addImpliedTags(expression)
		if err != nil {
			return nil, err
		}
	}

	relPath := storage.relPath(path)
	files, err := storage.Db.QueryFiles(expression, relPath)
	storage.absPaths(files)
	return files, err
}

// Retrieves the sets of duplicate files within the database.
func (storage *Storage) DuplicateFiles() ([]entities.Files, error) {
	fileSets, err := storage.Db.DuplicateFiles()

	for _, fileSet := range fileSets {
		storage.absPaths(fileSet)
	}

	return fileSets, err
}

// Adds a file to the database.
func (storage *Storage) AddFile(path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := storage.Db.InsertFile(relPath, fingerprint, modTime, size, isDir)
	storage.absPath(file)

	return file, err
}

// Updates a file in the database.
func (storage *Storage) UpdateFile(fileId entities.FileId, path string, fingerprint fingerprint.Fingerprint, modTime time.Time, size int64, isDir bool) (*entities.File, error) {
	relPath := storage.relPath(path)
	file, err := storage.Db.UpdateFile(fileId, relPath, fingerprint, modTime, size, isDir)
	storage.absPath(file)

	return file, err
}

// Deletes a file from the database.
func (storage *Storage) DeleteFile(fileId entities.FileId) error {
	return storage.Db.DeleteFile(fileId)
}

// Deletes a file if it is untagged
func (storage *Storage) DeleteFileIfUntagged(fileId entities.FileId) error {
	count, err := storage.FileTagCountByFileId(fileId, true)
	if err != nil {
		return err
	}
	if count == 0 {
		if err := storage.DeleteFile(fileId); err != nil {
			return err
		}
	}

	return nil
}

// Deletes the specified files if they are untagged
func (storage *Storage) DeleteUntaggedFiles(fileIds entities.FileIds) error {
	return storage.Db.DeleteUntaggedFiles(fileIds)
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

func (storage *Storage) addImpliedTags(expression query.Expression) (query.Expression, error) {
	implications, err := storage.Implications()
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

	return addImpliedTagsRecursive(expression, impliersByTag), nil
}

func addImpliedTagsRecursive(expression query.Expression, impliersByTag map[string][]string) query.Expression {
	switch typedExpression := expression.(type) {
	case query.OrExpression:
		typedExpression.LeftOperand = addImpliedTagsRecursive(typedExpression.LeftOperand, impliersByTag)
		typedExpression.RightOperand = addImpliedTagsRecursive(typedExpression.RightOperand, impliersByTag)
		return typedExpression
	case query.AndExpression:
		typedExpression.LeftOperand = addImpliedTagsRecursive(typedExpression.LeftOperand, impliersByTag)
		typedExpression.RightOperand = addImpliedTagsRecursive(typedExpression.RightOperand, impliersByTag)
		return typedExpression
	case query.NotExpression:
		typedExpression.Operand = addImpliedTagsRecursive(typedExpression.Operand, impliersByTag)
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
