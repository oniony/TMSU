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

package filesystem

import (
	"fmt"
	"github.com/oniony/TMSU/common/log"
	"os"
	"path/filepath"
)

type FileSystemFile struct {
	Path  string
	IsDir bool
}

func Enumerate(paths ...string) ([]FileSystemFile, error) {
	resultFiles := make([]FileSystemFile, 0, len(paths)*5)

	for _, path := range paths {
		var err error
		resultFiles, err = enumerate(path, resultFiles)
		if err != nil {
			return nil, err
		}
	}

	return resultFiles, nil
}

func EnumeratePaths(paths ...string) ([]string, error) {
	resultFiles, err := Enumerate(paths...)
	if err != nil {
		return nil, err
	}

	resultPaths := make([]string, len(resultFiles))
	for index, resultFile := range resultFiles {
		resultPaths[index] = resultFile.Path
	}

	return resultPaths, nil
}

func enumerate(path string, files []FileSystemFile) ([]FileSystemFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return files, nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied", path)
			return files, nil
		default:
			return nil, fmt.Errorf("%v: could not stat: %v", path, err)
		}
	}

	files = append(files, FileSystemFile{path, stat.IsDir()})

	if stat.IsDir() {
		dir, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not open directory: %v", path, err)
		}

		names, err := dir.Readdirnames(0)
		dir.Close()
		if err != nil {
			return nil, fmt.Errorf("%v: could not read directory entries: %v", path, err)
		}

		for _, name := range names {
			childPath := filepath.Join(path, name)
			files, err = enumerate(childPath, files)
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}
