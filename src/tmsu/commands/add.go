package main

import (
           "errors"
           "fmt"
           "log"
           "path/filepath"
       )

type AddCommand struct {}

func (this AddCommand) Name() string {
    return "add"
}

func (this AddCommand) Summary() string {
    return "adds a file without applying any tags"
}

func (this AddCommand) Help() string {
    return `  tmsu add FILE...

Adds the files specified without applying any tags.`
}

func (this AddCommand) Exec(args []string) error {
    if len(args) < 1 { return errors.New("At least one file to add must be specified.") }

    error := this.addPaths(args)
    if error != nil { return error }

    return nil
}

// implementation

func (this AddCommand) addPaths(paths []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    for _, path := range paths {
        _, _, error = this.addPath(db, path)

        if error != nil { return error }
    }

    return nil
}

func (this AddCommand) addPath(db *Database, path string) (*File, *FilePath, error) {
    absPath, error := filepath.Abs(path)
    if error != nil { return nil, nil, error }

    fingerprint, error := Fingerprint(absPath)
    if error != nil { return nil, nil, error }

    file, error := db.FileByFingerprint(fingerprint)
    if error != nil { return nil, nil, error }
    uniqueContents := file == nil

    filePath, error := db.FilePathByPath(absPath)
    if error != nil { return nil, nil, error }
    uniquePath := filePath == nil

    switch {
        case uniquePath && uniqueContents:
            fmt.Printf("Adding new file '%v'.\n", path)

            file, error = db.AddFile(fingerprint)
            if error != nil { return nil, nil, error }

            filePath, error = db.AddFilePath(file.Id, absPath)
            if error != nil { return nil, nil, error }
        case uniquePath:
            fmt.Printf("Adding file '%v'.\n", path)

            filePath, error = db.AddFilePath(file.Id, absPath)
            if error != nil { return nil, nil, error }
        case uniqueContents:
            fmt.Printf("Updating with modified file '%v'.\n", path)

            //TODO contents have changed, update file-path
            log.Fatalf("Not implemented.")
    }

    return file, filePath, nil
}
