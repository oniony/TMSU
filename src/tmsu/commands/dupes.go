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
	"path/filepath"
	"tmsu/common"
	"tmsu/database"
)

type DupesCommand struct{}

func (DupesCommand) Name() string {
	return "dupes"
}

func (DupesCommand) Synopsis() string {
	return "Identify duplicate files"
}

func (DupesCommand) Description() string {
	return `tmsu dupes [FILE]...

Identifies all files in the database that are exact duplicates of FILE. If no
FILE is specified then identifies duplicates between files in the database.`
}

func (command DupesCommand) Exec(args []string) error {
	switch len(args) {
	case 0:
		command.findDuplicatesInDb()
	default:
		return command.findDuplicatesOf(args)
	}

	return nil
}

func (DupesCommand) findDuplicatesInDb() error {
	db, err := database.OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	fileSets, err := db.DuplicateFiles()
	if err != nil {
		return err
	}

	for index, fileSet := range fileSets {
		if index > 0 {
			fmt.Println()
		}

		fmt.Printf("Set of %v duplicates:\n", len(fileSet))

		for _, file := range fileSet {
			relPath := common.MakeRelative(file.Path())
			fmt.Printf("  %v\n", relPath)
		}
	}

	return nil
}

func (DupesCommand) findDuplicatesOf(paths []string) error {
	db, err := database.OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	first := true
	for _, path := range paths {
		fingerprint, err := common.Fingerprint(path)
		if err != nil {
			return err
		}

		files, err := db.FilesByFingerprint(fingerprint)
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// filter out the file we're searching on
		dupes := where(files, func(file database.File) bool { return file.Path() != absPath })

		if len(paths) > 1 && len(dupes) > 0 {
			if first {
				first = false
			} else {
				fmt.Println()
			}

			fmt.Printf("%v duplicates of %v:\n", len(dupes), path)

			for _, dupe := range dupes {
				relPath := common.MakeRelative(dupe.Path())
				fmt.Printf("  %v\n", relPath)
			}
		} else {
			for _, dupe := range dupes {
				relPath := common.MakeRelative(dupe.Path())
				fmt.Println(relPath)
			}
		}
	}

	return nil
}

func where(files []database.File, predicate func(database.File) bool) []database.File {
	result := make([]database.File, 0, len(files))

	for _, file := range files {
		if predicate(file) {
			result = append(result, file)
		}
	}

	return result
}
