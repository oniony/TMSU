/*
Copyright 2011-2015 Paul Ruane.

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

package path

import (
	"os"
	"path/filepath"
	"strings"
)

type FileSystemFile struct {
	Path  string
	IsDir bool
}

func Rel(path string) string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return path
	}

    return RelTo(path, workingDirectory)
}

func RelTo(path, to string) string {
    var err error

    path, err = filepath.Abs(path)
    if err != nil {
        panic("could not get absolute path")
    }

    to, err = filepath.Abs(to)
    if err != nil {
        panic("could not get absolute path")
    }

	if path == to {
		return "."
	}

    prefix := trailingSeparator(to)
	if strings.HasPrefix(path, prefix) {
	    // can't use filepath.Join as it strips the leading './'
        return "." + string(filepath.Separator) + path[len(prefix):]
	}

    to = filepath.Dir(to)
    prefix = trailingSeparator(to)
    if strings.HasPrefix(path, prefix) {
	    // can't use filepath.Join as it strips the leading './'
        return ".." + string(filepath.Separator) + path[len(prefix):]
    }

	return path
}

// unexported

func trailingSeparator(path string) string {
    if path[len(path)-1] == filepath.Separator {
        return path
    }

    return path + string(filepath.Separator)
}
