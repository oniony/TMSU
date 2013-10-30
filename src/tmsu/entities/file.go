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

package entities

import (
	"path/filepath"
	"time"
	"tmsu/fingerprint"
)

type File struct {
	Id          uint
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
	FileId    uint
	Directory string
	Name      string
	TagCount  uint
}
