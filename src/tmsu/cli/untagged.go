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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"tmsu/common/filesystem"
	_path "tmsu/common/path"
	"tmsu/storage"
)

var UntaggedCommand = Command{
	Name:     "untagged",
	Synopsis: "List untagged files in the filesystem",
	Description: `tmsu untagged [PATH]...

Identify untagged files in the filesystem.  

Where PATHs are not specified, untagged items under the current working
directory are shown.

Examples:

    $ tmsu untagged
    $ tmsu untagged /home/fred/drawings`,
	Options: Options{Option{"--recursive", "-r", "recursively identify untagged items within directories", false, ""}},
	Exec:    untaggedExec,
}

func untaggedExec(options Options, args []string) error {
	recursive := options.HasOption("--recursive")

	paths := args
	if len(paths) == 0 {
		var err error
		paths, err = workingDirectoryEntries()
		if err != nil {
			return err
		}
	}

	if recursive {
		var err error
		paths, err = filesystem.EnumeratePaths(paths...)
		if err != nil {
			return fmt.Errorf("could not enumerate paths: %v", err)
		}
	}

	untaggedAbsPaths, err := findUntagged(paths)
	if err != nil {
		return err
	}

	untaggedRelPaths := make([]string, len(untaggedAbsPaths))
	for index, untaggedAbsPath := range untaggedAbsPaths {
		untaggedRelPaths[index] = _path.Rel(untaggedAbsPath)
	}

	sort.Strings(untaggedRelPaths)

	for _, untaggedRelPath := range untaggedRelPaths {
		fmt.Println(untaggedRelPath)
	}

	return nil
}

func findUntagged(paths []string) ([]string, error) {
	untaggedPaths := make([]string, 0, 10)

	store, err := storage.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return nil, fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file == nil {
			untaggedPaths = append(untaggedPaths, absPath)
		}
	}

	return untaggedPaths, nil
}

func workingDirectoryEntries() ([]string, error) {
	wd, err := os.Open(".")
	if err != nil {
		return nil, fmt.Errorf("could not open working directory", err)
	}

	names, err := wd.Readdirnames(0)
	wd.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read working directory entries", err)
	}

	return names, nil
}
