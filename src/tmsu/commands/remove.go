package main

import (
           "errors"
           "path/filepath"
       )

type RemoveCommand struct {}


func (this RemoveCommand) Name() string {
    return "remove"
}

func (this RemoveCommand) Description() string {
    return "removes a previously added file"
}

func (this RemoveCommand) Exec(args []string) error {
    if len(args) < 1 { return errors.New("At least one file to remove must be specified.") }

    error := this.removePaths(args)
    if error != nil { return error }

    return nil
}

// implementation

func (this RemoveCommand) removePaths(paths []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    for _, path := range paths {
        error = this.removePath(db, path)
        if error != nil { return error }
    }

    return nil
}

func (this RemoveCommand) removePath(db *Database, path string) error {
    absPath, error := filepath.Abs(path)
    if error != nil { return error }

    error = db.DeleteFilePath(absPath)
    if error != nil { return error }

    return nil
}
