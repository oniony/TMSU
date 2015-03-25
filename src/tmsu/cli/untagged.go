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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"tmsu/common/log"
	_path "tmsu/common/path"
	"tmsu/storage"
)

var UntaggedCommand = Command{
	Name:     "untagged",
	Synopsis: "List untagged files",
	Usages:   []string{"tmsu untagged [OPTION]... [PATH]..."},
	Description: `Identify untagged files in the filesystem.  

Where PATHs are not specified, untagged items under the current working directory are shown.`,
	Examples: []string{"$ tmsu untagged",
		"$ tmsu untagged /home/fred/drawings"},
	Options: Options{Option{"--directory", "-d", "do not examine directory contents (non-recursive)", false, ""}},
	Exec:    untaggedExec,
}

func untaggedExec(store *storage.Storage, options Options, args []string) error {
	recursive := !options.HasOption("--directory")

	paths := args
	if len(paths) == 0 {
		var err error
		paths, err = directoryEntries(".")
		if err != nil {
			return err
		}
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	if err := findUntagged(store, tx, paths, recursive); err != nil {
		return err
	}

	return nil
}

func findUntagged(store *storage.Storage, tx *storage.Tx, paths []string, recursive bool) error {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		//TODO PERF no need to retrieve file: we merely need to know it exists
		file, err := store.FileByPath(tx, absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file == nil {
			relPath := _path.Rel(absPath)
			fmt.Println(relPath)
		}

		if recursive {
			entries, err := directoryEntries(path)
			if err != nil {
				return err
			}

			findUntagged(store, tx, entries, true)
		}
	}

	return nil
}

func directoryEntries(path string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Warnf("%v: does not exist", path)
			return []string{}, nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied", path)
			return []string{}, nil
		default:
			return nil, fmt.Errorf("%v: could not stat", path, err)
		}
	}

	if !stat.IsDir() {
		return []string{}, nil
	}

	dir, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%v could not open directory", path, err)
	}

	names, err := dir.Readdirnames(0)
	dir.Close()
	if err != nil {
		return nil, fmt.Errorf("%v: could not read directory entries", path, err)
	}

	entries := make([]string, len(names))
	for index, name := range names {
		entries[index] = filepath.Join(path, name)
	}

	return entries, nil
}
