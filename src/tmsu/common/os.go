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

package common

import (
	"os"
	"path/filepath"
)

func IsRegular(fileInfo os.FileInfo) bool {
	return fileInfo.Mode()&os.ModeType == 0
}

func DirectoryEntries(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entryNames, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	entryPaths := make([]string, len(entryNames))
	for index, entryName := range entryNames {
		entryPaths[index] = filepath.Join(path, entryName)
	}

	return entryPaths, nil
}
