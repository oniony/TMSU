// Copyright 2011-2017 Paul Ruane.

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

package entities

import (
	"github.com/oniony/TMSU/common/fingerprint"
	"path/filepath"
	"sort"
	"time"
)

type FileId uint

type FileIds []FileId

func (fileIds FileIds) Len() int {
	return len(fileIds)
}

func (fileIds FileIds) Less(i, j int) bool {
	return fileIds[i] < fileIds[j]
}

func (fileIds FileIds) Swap(i, j int) {
	fileIds[i], fileIds[j] = fileIds[j], fileIds[i]
}

func (fileIds FileIds) Uniq() FileIds {
	if len(fileIds) == 0 {
		return fileIds
	}

	sort.Sort(fileIds)
	uniq := FileIds{fileIds[0]}
	previous := fileIds[0]

	for _, fileId := range fileIds[1:] {
		if fileId != previous {
			uniq = append(uniq, fileId)
			previous = fileId
		}
	}

	return uniq
}

type File struct {
	Id          FileId
	Directory   string
	Name        string
	Fingerprint fingerprint.Fingerprint
	ModTime     time.Time
	Size        int64
	IsDir       bool
}

func (file File) Path() string {
	return filepath.Join(file.Directory, file.Name)
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

type FileTagCount struct {
	FileId    FileId
	Directory string
	Name      string
	TagCount  uint
}
