package main

import (
         "errors"
         "fmt"
       )

type UntagCommand struct {}

func (this UntagCommand) Name() string {
    return "untag"
}

func (this UntagCommand) Summary() string {
    return "removes tags from a file"
}

func (this UntagCommand) Help() string {
    return `  tmsu untag FILE TAG...

Disassociates the specified file FILE with the tag(s) specified.`
}

func (this UntagCommand) Exec(args []string) error {
    if len(args) < 2 { return errors.New("File to untag and tags to remove must be specified.") }

    error := this.untagPath(args[0], args[1:])
    if error != nil { return error }

    return nil
}

// implementation

func (this UntagCommand) untagPath(path string, tagNames []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    file, error := db.FileByPath(path)
    if error != nil { return error }

    for _, tagName := range tagNames {
        error = this.unapplyTag(db, path, file.Id, tagName)
        if error != nil { return error }

        //TODO remove the file if has no tags
    }

    return nil
}

func (this UntagCommand) unapplyTag(db *Database, path string, fileId uint, tagName string) error {
    tag, error := db.TagByName(tagName)
    if error != nil { return error }
    if tag == nil { errors.New("No such tag" + tagName) }

    fileTag, error := db.FileTagByFileAndTag(fileId, tag.Id)
    if error != nil { return error }
    if fileTag == nil { errors.New(fmt.Sprintf("File '%v' is not tagged '%v'.", path, tagName)) }

    if fileTag != nil {
        error := db.RemoveFileTag(fileId, tag.Id)
        if error != nil { return error }

        fmt.Printf("Untagged file '%v' with '%v'.\n", path, tagName)
    } else {
        fmt.Printf("File '%v' is not tagged '%v'.\n", path, tagName)
    }

    return nil
}
