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
package main
*/

package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

type DupesCommand struct{}

func (this DupesCommand) Name() string {
	return "dupes"
}

func (this DupesCommand) Summary() string {
	return "identifies any duplicate files"
}

func (this DupesCommand) Help() string {
	return `  tmsu dupes [FILE]

Identifies all files in the database that are exact duplicates of FILE.

When FILE is omitted duplicates within the database are identified.`
}

func (this DupesCommand) Exec(args []string) error {
    argCount := len(args)
    if argCount > 1 { errors.New("Only a single file can be specified.") }

    switch argCount {
        case 0: return this.findDuplicates()
        case 1: return this.findDuplicatesOf(args[0])
    }

	return nil
}

// implementation

func (this DupesCommand) findDuplicates() error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    fileSets, error := db.DuplicateFiles()
    if error != nil { return error }

    for index, fileSet := range fileSets {
        if index > 0 { fmt.Println() }

        for _, file := range fileSet {
            fmt.Println(file.Path)
        }
    }

    return nil
}

func (this DupesCommand) findDuplicatesOf(path string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    fingerprint, error := Fingerprint(path)
    if error != nil { return error }

    files, error := db.FilesByFingerprint(fingerprint)
    if error != nil { return error }

    absPath, error := filepath.Abs(path)
    if error != nil { return error }

    for _, file := range files {
        if file.Path == absPath { continue }

        fmt.Println(file.Path)
    }

    return nil
}
