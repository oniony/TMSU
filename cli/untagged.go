// Paul Ruane.

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
	"github.com/oniony/TMSU/common/log"
	_path "github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
)

var UntaggedCommand = Command{
	Name:     "untagged",
	Synopsis: "List untagged files",
	Usages:   []string{"tmsu untagged [OPTION]... [PATH]..."},
	Description: `Identify untagged files in the filesystem.  

Where PATHs are not specified, untagged items under the current working directory are shown.`,
	Examples: []string{"$ tmsu untagged",
		"$ tmsu untagged /home/fred/drawings"},
	Options: Options{Option{"--directory", "-d", "do not examine directory contents (non-recursive)", false, ""},
		Option{"--count", "-c", "list the number of files rather than their names", false, ""},
		Option{"--no-dereference", "-P", "do not dereference symbolic links", false, ""}},
	Exec: untaggedExec,
}

// unexported

func untaggedExec(options Options, args []string, databasePath string) (error, warnings) {
	recursive := !options.HasOption("--directory")
	count := options.HasOption("--count")
	followSymlinks := !options.HasOption("--no-dereference")

	paths := args
	if len(paths) == 0 {
		var err error
		paths, err = directoryEntries(".")
		if err != nil {
			return err, nil
		}
	}

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	if count {
		count, err := findUntaggedCount(store, tx, paths, recursive, followSymlinks)
		if err != nil {
			return err, nil
		}

		fmt.Println(count)
	} else {
		if err := findUntagged(store, tx, paths, recursive, followSymlinks); err != nil {
			return err, nil
		}
	}

	return nil, nil
}

func findUntagged(store *storage.Storage, tx *storage.Tx, paths []string, recursive, followSymlinks bool) error {
	var action = func(absPath string) {
		relPath := _path.Rel(absPath)
		fmt.Println(relPath)
	}

	return findUntaggedFunc(store, tx, paths, recursive, followSymlinks, action)
}

func findUntaggedCount(store *storage.Storage, tx *storage.Tx, paths []string, recursive, followSymlinks bool) (uint, error) {
	var count uint

	var action = func(absPath string) {
		count++
	}

	err := findUntaggedFunc(store, tx, paths, recursive, followSymlinks, action)

	return count, err
}

func findUntaggedFunc(store *storage.Storage, tx *storage.Tx, paths []string, recursive, followSymlinks bool, action func(absPath string)) error {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		if followSymlinks {
			log.Infof(2, "%v: resolving path", path)

			absPath, err = _path.Dereference(absPath)
			if err != nil {
				return fmt.Errorf("%v: could not dereference path: %v", path, err)
			}
		}

		//TODO PERF no need to retrieve file: we merely need to know it exists
		file, err := store.FileByPath(tx, absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file == nil {
			action(absPath)
		}

		if recursive {
			entries, err := directoryEntries(path)
			if err != nil {
				return err
			}

			findUntaggedFunc(store, tx, entries, true, followSymlinks, action)
		}
	}

	return nil
}

func directoryEntries(path string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			//TODO return as warning
			log.Warnf("%v: does not exist", path)
			return []string{}, nil
		case os.IsPermission(err):
			//TODO return as warning
			log.Warnf("%v: permission denied", path)
			return []string{}, nil
		default:
			return nil, fmt.Errorf("%v: could not stat: %v", path, err)
		}
	}

	if !stat.IsDir() {
		return []string{}, nil
	}

	dir, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%v could not open directory: %v", path, err)
	}

	names, err := dir.Readdirnames(0)
	dir.Close()
	if err != nil {
		return nil, fmt.Errorf("%v: could not read directory entries: %v", path, err)
	}

	entries := make([]string, len(names))
	for index, name := range names {
		entries[index] = filepath.Join(path, name)
	}

	return entries, nil
}
