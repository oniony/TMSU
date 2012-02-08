/*
Copyright 2011 Paul Ruane.

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

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"tmsu/core"
	"tmsu/database"
)

type StatusCommand struct{}

func (StatusCommand) Name() string {
	return "status"
}

func (StatusCommand) Synopsis() string {
	return "List file status"
}

func (StatusCommand) Description() string {
	return `tmsu status [FILE]...

Shows the tag status of files.

Where one or more FILEs are specified, the status of these files is shown.
Where FILE is a directory, details of all files within the specified directory
and its descendent directories are shown.

Where no FILE is specified, details of all files within the current directory
and its descendent directories are shown.`
}

func (command StatusCommand) Exec(args []string) error {
	tagged := make([]string, 0, 10)
	untagged := make([]string, 0, 10)
	missing := make([]string, 0, 10)
	allFiles := false
	var err error

	if len(args) > 0 && args[0] == "--all" {
		allFiles = true
		args = args[1:]
	}

	if len(args) == 0 {
		tagged, untagged, missing, err = command.status([]string {"."}, tagged, untagged, missing, allFiles)
	} else {
		tagged, untagged, missing, err = command.status(args, tagged, untagged, missing, allFiles)
	}

	if err != nil {
		return err
	}

	for _, path := range tagged {
		fmt.Printf("T %v\n", path)
	}

	for _, path := range missing {
		fmt.Printf("! %v\n", path)
	}

	for _, path := range untagged {
		fmt.Printf("? %v\n", path)
	}

	return nil
}

func (command StatusCommand) status(paths []string, tagged []string, untagged []string, missing []string, allFiles bool) ([]string, []string, []string, error) {
	for _, path := range paths {
        absPath, err := filepath.Abs(path)
        if err != nil {
            return nil, nil, nil, err
        }

		databaseEntries, err := command.getDatabaseEntries(absPath)
		if err != nil {
			return nil, nil, nil, err
		}

		fileSystemEntries, err := command.getFileSystemEntries(absPath, allFiles)
		if err != nil {
			return nil, nil, nil, err
		}

		for _, entryPath := range databaseEntries {
		    fmt.Println("Considering", entryPath)
			if contains(fileSystemEntries, entryPath) {
			    relPath := core.MakeRelative(entryPath)
				tagged = append(tagged, relPath)
			} else {
			    relPath := core.MakeRelative(entryPath)
				missing = append(missing, relPath)
			}
		}

		for _, entryPath := range fileSystemEntries {
			if !contains(databaseEntries, entryPath) {
			    relPath := core.MakeRelative(entryPath)
				untagged = append(untagged, relPath)
			}
		}
	}

	return tagged, untagged, missing, nil
}

func (command StatusCommand) getFileSystemEntries(path string, allFiles bool) ([]string, error) {
	return command.getFileSystemEntriesRecursive(path, make([]string, 0, 10), allFiles)
}

func (command StatusCommand) getFileSystemEntriesRecursive(path string, entries []string, allFiles bool) ([]string, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	basename := filepath.Base(path)

	if basename[0] != '.' || allFiles {
		if core.IsRegular(fileInfo) {
			entries = append(entries, path)
		} else if fileInfo.IsDir() {
			entries = append(entries, path)

			childEntries, err := core.DirectoryEntries(path)
			if err != nil {
				switch terr := err.(type) {
				case *os.PathError:
					switch terr.Err {
					case os.EACCES:
						core.Warn("'%v': permission denied.", path)
					default:
						return nil, err
					}
				default:
					return nil, err
				}
			}

			for _, entry := range childEntries {
				entries, err = command.getFileSystemEntriesRecursive(entry, entries, allFiles)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return entries, nil
}

func (StatusCommand) getDatabaseEntries(path string) ([]string, error) {
	db, err := database.OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	files, err := db.FilesByDirectory(path)
	if err != nil {
		return nil, err
	}

	entries := make([]string, 0, len(files))
	for _, file := range files {
		entries = append(entries, file.Path())
	}

	return entries, nil
}

func contains(strings []string, find string) bool {
	for _, str := range strings {
		if str == find {
			return true
		}
	}

	return false
}
