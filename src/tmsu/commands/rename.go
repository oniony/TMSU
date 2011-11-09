package main

import (
           "errors"
           "log"
       )

type RenameCommand struct {}

func (this RenameCommand) Name() string {
    return "rename"
}

func (this RenameCommand) Description() string {
    return "renames a tag"
}

func (this RenameCommand) Exec(args []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    currentName := args[0]
    newName := args[1]

    tag, error := db.TagByName(currentName)
    if error != nil { return error }
    if tag == nil { errors.New("No such tag '" + currentName + "'.") }

    newTag, error := db.TagByName(newName)
    if error != nil { return error }

    //TODO only merge if flag set

    if newTag == nil {
        _, error = db.RenameTag(tag.Id, newName)
        if error != nil { return error }
    } else {
        error = db.MigrateFileTags(tag.Id, newTag.Id)
        if error != nil { return error }

        error = db.DeleteTag(tag.Id)
        if error != nil { return error }
    }

    return nil
}
