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

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"tmsu/common"
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

type StatusReport struct {
	Tagged   []string
	Modified []string
	Missing  []string
	Untagged []string
}

func NewReport() *StatusReport {
	return &StatusReport{make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10)}
}

func (command StatusCommand) Exec(args []string) error {
	allFiles := false
	if len(args) > 0 && args[0] == "--all" {
		allFiles = true
		args = args[1:]
	}

	var err error
	report := NewReport()
	if len(args) == 0 {
		entries, err := common.DirectoryEntries(".")
		if err != nil {
			return err
		}

		report, err = command.status(entries, report, allFiles)
	} else {
		report, err = command.status(args, report, allFiles)
	}

	if err != nil {
		return err
	}

	for _, path := range report.Tagged {
		fmt.Println("T", path)
	}

	for _, path := range report.Modified {
		fmt.Println("M", path)
	}

	for _, path := range report.Missing {
		fmt.Println("!", path)
	}

	for _, path := range report.Untagged {
		fmt.Println("?", path)
	}

	return nil
}

func (command StatusCommand) status(paths []string, report *StatusReport, allFiles bool) (*StatusReport, error) {
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}

		databaseEntries, err := command.getDatabaseEntries(absPath)
		if err != nil {
			return nil, err
		}

		fileSystemEntries, err := command.getFileSystemEntries(absPath, allFiles)
		if err != nil {
			return nil, err
		}

		for _, entry := range databaseEntries {
			relPath := common.MakeRelative(entry.Path())

			if contains(fileSystemEntries, entry.Path()) {
				fingerprint, err := common.Fingerprint(entry.Path())
				if err != nil {
					return nil, err
				}

				if fingerprint != entry.Fingerprint {
					report.Modified = append(report.Modified, relPath)
				} else {
					report.Tagged = append(report.Tagged, relPath)
				}
			} else {
				relPath := common.MakeRelative(entry.Path())
				report.Missing = append(report.Missing, relPath)
			}
		}

		for _, entryPath := range fileSystemEntries {
			if _, contains := databaseEntries[entryPath]; !contains {
				relPath := common.MakeRelative(entryPath)
				report.Untagged = append(report.Untagged, relPath)
			}
		}
	}

	return report, nil
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
		if common.IsRegular(fileInfo) {
			entries = append(entries, path)
		} else if fileInfo.IsDir() {
			entries = append(entries, path)

			childEntries, err := common.DirectoryEntries(path)
			if err != nil {
				if os.IsPermission(err) {
					common.Warnf("'%v': permission denied.", path)
				} else {
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

func (StatusCommand) getDatabaseEntries(path string) (map[string]*database.File, error) {
	db, err := database.OpenDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	files, err := db.FilesByDirectory(path)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]*database.File)
	for _, file := range files {
		entries[file.Path()] = file
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
