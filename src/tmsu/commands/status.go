package main

//TODO show modified files

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
    var error error

    if len(args) == 0 {
        tagged, untagged, error = this.statusChildren(".", tagged, untagged)
    } else {
        tagged, untagged, error = this.status(args, tagged, untagged)
    }

    if error != nil {
        return error
    }

    for _, path := range tagged {
        fmt.Printf("T %v\n", path)
    }

    for _, path := range untagged {
        fmt.Printf("U %v\n", path)
    }

    return nil
}

func (this StatusCommand) status(paths []string, tagged []string, untagged []string) ([]string, []string, error) {
    db, error := OpenDatabase(databasePath())
    if error != nil {
        return nil, nil, error
    }

    for _, path := range paths {
        absPath, error := filepath.Abs(path)
        if error != nil {
            return nil, nil, error
        }

        fileInfo, error := os.Stat(absPath)
        if error != nil {
            return nil, nil, error
        }

        if fileInfo.IsRegular() || fileInfo.IsSymlink() {
            file, error := db.FileByPath(absPath)
            if error != nil {
                return nil, nil, error
            }

            //TODO show file type (dir, reg, lnk) &c
            if file == nil {
                untagged = append(untagged, absPath)
            } else {
                tagged = append(tagged, absPath)
            }
        } else if fileInfo.IsDirectory() {
            tagged, untagged, error = this.statusChildren(absPath, tagged, untagged)
            if error != nil {
                return nil, nil, error
            }
        }
    }

    return tagged, untagged, nil
}

func (this StatusCommand) statusChildren(path string, tagged []string, untagged []string) ([]string, []string, error) {
    file, error := os.Open(path)
    if error != nil {
        return nil, nil, error
    }
    defer file.Close()

    dirNames, error := file.Readdirnames(0)
    if error != nil {
        return nil, nil, error
    }

    childPaths := make([]string, len(dirNames))
    for index, dirName := range dirNames {
        childPaths[index] = filepath.Join(path, dirName)
    }

    tagged, untagged, error = this.status(childPaths, tagged, untagged)
    if error != nil {
        return nil, nil, error
    }

    return tagged, untagged, nil
}

