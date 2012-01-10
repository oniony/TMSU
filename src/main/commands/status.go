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
    "strings"
)

type StatusCommand struct {}

func (StatusCommand) Name() string {
    return "status"
}

func (StatusCommand) Summary() string {
    return "lists file status"
}

func (StatusCommand) Help() string {
    return `tmsu status
tmsu status FILE...

Shows the status of files.`
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
        tagged, untagged, missing, err = command.status([]string { "." }, tagged, untagged, missing, allFiles)
    } else {
        tagged, untagged, missing, err = command.status(args, tagged, untagged, missing, allFiles)
    }

    if err != nil { return err }

    for _, absPath := range tagged {
        path, err := makeRelative(absPath)
        if err != nil { return err }

        fmt.Printf("T %v\n", path)
    }

    for _, absPath := range missing {
        path, err := makeRelative(absPath)
        if err != nil { return err }

        fmt.Printf("! %v\n", path)
    }

    for _, absPath := range untagged {
        path, err := makeRelative(absPath)
        if err != nil { return err }

        fmt.Printf("? %v\n", path)
    }

    return nil
}

func (command StatusCommand) status(paths []string, tagged []string, untagged []string, missing []string, allFiles bool) ([]string, []string, []string, error) {
    for _, path := range paths {
        databaseEntries, err := command.getDatabaseEntries(path)
        if err != nil { return nil, nil, nil, err }

        fileSystemEntries, err := command.getFileSystemEntries(path, allFiles)
        if err != nil { return nil, nil, nil, err }

        for _, entry := range databaseEntries {
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

func (command StatusCommand) getFileSystemEntries(path string, allFiles bool) ([]string, error) {
    return command.getFileSystemEntriesRecursive(path, make([]string, 0, 10), allFiles)
}

func (command StatusCommand) getFileSystemEntriesRecursive(path string, entries []string, allFiles bool) ([]string, error) {
    fileInfo, err := os.Lstat(path)
    if err != nil { return nil, err }

    absPath, err := filepath.Abs(path)
    if err != nil { return nil, err }

    basename := filepath.Base(absPath)

    if basename[0] != '.' || allFiles {
        if isRegular(fileInfo)  {
            entries = append(entries, absPath)
        } else if fileInfo.IsDir() {
            childEntries, err := directoryEntries(absPath)
            if err != nil { return nil, err }

            for _, entry := range childEntries {
                entries, err = command.getFileSystemEntriesRecursive(entry, entries, allFiles)
                if err != nil { return nil, err }
            }
        }
    }

    return entries, nil
}

func (StatusCommand) getDatabaseEntries(path string) ([]string, error) {
    db, err := OpenDatabase()
    if err != nil { return nil, err }
    defer db.Close()

    absPath, err := filepath.Abs(path)
    if err != nil { return nil, err }

    files, err := db.FilesByDirectory(absPath)
    if err != nil { return nil, err }

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
    workingDirectory, err := os.Getwd()
    if err != nil { return "", err }

    workingDirectory += string(filepath.Separator)

    if strings.HasPrefix(path, workingDirectory) {
        return path[len(workingDirectory):], nil
    }

    return path, nil
}
