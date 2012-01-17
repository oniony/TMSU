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

package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

type DupesCommand struct{}

func (DupesCommand) Name() string {
	return "dupes"
}

func (DupesCommand) Summary() string {
	return "identifies any duplicate files"
}

func (DupesCommand) Help() string {
	return `tmsu dupes [FILE]

Identifies all files in the database that are exact duplicates of FILE.

When FILE is omitted duplicates within the database are identified.`
}

func (command DupesCommand) Exec(args []string) error {
	argCount := len(args)
	if argCount > 1 {
		errors.New("Only a single file can be specified.")
	}

	switch argCount {
	case 0:
		return command.findDuplicates()
	case 1:
		return command.findDuplicatesOf(args[0])
	}

	return nil
}

func (DupesCommand) findDuplicates() error {
	db, err := OpenDatabase()
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

		fmt.Printf("%v duplicate files:\n", len(fileSet))

		for _, file := range fileSet {
			fmt.Printf("  %v\n", file.Path())
		}
	}

	return nil
}

func (DupesCommand) findDuplicatesOf(path string) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	fingerprint, err := Fingerprint(path)
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

	for _, file := range files {
		if file.Path() == absPath {
			continue
		}

		fmt.Println(file.Path())
	}

	return nil
}
