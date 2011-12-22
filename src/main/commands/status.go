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
    "path/filepath"
    "fmt"
    "os"
)

type StatusCommand struct {}

func (this StatusCommand) Name() string {
    return "status"
}

func (this StatusCommand) Summary() string {
    return "lists file status"
}

func (this StatusCommand) Help() string {
    return `tmsu status
tmsu status FILE...

Shows the status of files.`
}

func (this StatusCommand) Exec(args []string) error {
    tagged := make([]string, 0, 10)
    untagged := make([]string, 0, 10)
    missing := make([]string, 0, 10)
    var error error

    if len(args) == 0 {
        tagged, untagged, missing, error = this.status([]string { "." }, tagged, untagged, missing)
    } else {
        tagged, untagged, missing, error = this.status(args, tagged, untagged, missing)
    }

    if error != nil {
        return error
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

func (this StatusCommand) status(paths []string, tagged []string, untagged []string, missing []string) ([]string, []string, []string, error) {
    db, error := OpenDatabase(databasePath())
    if error != nil { return nil, nil, nil, error }
    defer db.Close()

    return this.statusRecursive(db, paths, tagged, untagged, missing)
}

func (this StatusCommand) statusRecursive(db *Database, paths []string, tagged []string, untagged []string, missing []string) ([]string, []string, []string, error) {
    for _, path := range paths {
        fileInfo, error := os.Lstat(path)
        if error != nil { return nil, nil, nil, error }

        absPath, error := filepath.Abs(path)
        if error != nil { return nil, nil, nil, error }

        if isRegular(fileInfo)  {
            file, error := db.FileByPath(absPath)
            if error != nil { return nil, nil, nil, error }

            if file == nil {
                untagged = append(untagged, path)
            } else {
                tagged = append(tagged, path)
            }
        } else if fileInfo.IsDir() {
            files, error := db.FilesByDirectory(absPath)
            if error != nil { return nil, nil, nil, error }

            for _, file := range files {
                _, error := os.Lstat(file.Path())

                if error != nil {
                    if error.(*os.PathError).Err == os.ENOENT {
                        missingFilePath := filepath.Join(path, file.Name)
                        missing = append(missing, missingFilePath)
                    } else {
                        return nil, nil, nil, error
                    }
                }
            }

            childPaths, error := directoryEntries(path)
            if error != nil { return nil, nil, nil, error }

            tagged, untagged, missing, error = this.statusRecursive(db, childPaths, tagged, untagged, missing)
            if error != nil { return nil, nil, nil, error }
        }
    }

    return tagged, untagged, missing, nil
}
