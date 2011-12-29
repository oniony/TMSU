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
    allFiles := false
    var error error

    if len(args) > 0 && args[0] == "--all" {
        allFiles = true
        args = args[1:]
    }

    if len(args) == 0 {
        tagged, untagged, missing, error = this.status([]string { "." }, tagged, untagged, missing, allFiles)
    } else {
        tagged, untagged, missing, error = this.status(args, tagged, untagged, missing, allFiles)
    }

    if error != nil { return error }

    for _, absPath := range tagged {
        path, error := makeRelative(absPath)
        if error != nil { return error }

        fmt.Printf("T %v\n", path)
    }

    for _, absPath := range missing {
        path, error := makeRelative(absPath)
        if error != nil { return error }

        fmt.Printf("! %v\n", path)
    }

    for _, absPath := range untagged {
        path, error := makeRelative(absPath)
        if error != nil { return error }

        fmt.Printf("? %v\n", path)
    }

    return nil
}

func (this StatusCommand) status(paths []string, tagged []string, untagged []string, missing []string, allFiles bool) ([]string, []string, []string, error) {
    for _, path := range paths {
        databaseEntries, error := this.getDatabaseEntries(path)
        if error != nil { return nil, nil, nil, error }

        fileSystemEntries, error := this.getFileSystemEntries(path, allFiles)
        if error != nil { return nil, nil, nil, error }

        for _, entry := range databaseEntries {
            fmt.Printf("Searching FS entries for '%v'\n", entry)
            if contains(fileSystemEntries, entry) {
                tagged = append(tagged, entry)
            } else {
                missing = append(missing, entry)
            }
        }

        for _, entry := range fileSystemEntries {
            if !contains(databaseEntries, entry) {
                untagged = append(untagged, entry)
            }
        }
    }

    return tagged, untagged, missing, nil
}

func (this StatusCommand) getFileSystemEntries(path string, allFiles bool) ([]string, error) {
    return this.getFileSystemEntriesRecursive(path, make([]string, 0, 10), allFiles)
}

func (this StatusCommand) getFileSystemEntriesRecursive(path string, entries []string, allFiles bool) ([]string, error) {
    fileInfo, error := os.Lstat(path)
    if error != nil { return nil, error }

    absPath, error := filepath.Abs(path)
    if error != nil { return nil, error }

    basename := filepath.Base(absPath)

    if basename[0] != '.' || allFiles {
        if isRegular(fileInfo)  {
            entries = append(entries, absPath)
        } else if fileInfo.IsDir() {
            childEntries, error := directoryEntries(absPath)
            if error != nil { return nil, error }

            for _, entry := range childEntries {
                entries, error = this.getFileSystemEntriesRecursive(entry, entries, allFiles)
                if error != nil { return nil, error }
            }
        }
    }

    return entries, nil
}

func (this StatusCommand) getDatabaseEntries(path string) ([]string, error) {
    db, error := OpenDatabase(databasePath())
    if error != nil { return nil, error }
    defer db.Close()

    absPath, error := filepath.Abs(path)
    if error != nil { return nil, error }

    files, error := db.FilesByDirectory(absPath)
    if error != nil { return nil, error }

    entries := make([]string, 0, len(files))
    for _, file := range files {
        entries = append(entries, file.Path())
    }

    return entries, nil
}

func contains(strings []string, find string) bool {
    for _, str := range strings {
        if str == find { return true }
    }

    return false
}

func makeRelative(path string) (string, error) {
    workingDirectory, error := os.Getwd()
    if error != nil { return "", error }

    relPath, error := filepath.Rel(workingDirectory, path)
    if error != nil { return path, nil }

    return relPath, nil
}
