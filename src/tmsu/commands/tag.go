package main

import (
         "errors"
         "fmt"
       )

type TagCommand struct {
    AddCommand
}

func (this TagCommand) Name() string {
    return "tag"
}

func (this TagCommand) Description() string {
    return "applies one or more tags to a file"
}

func (this TagCommand) Exec(args []string) error {
    if len(args) < 2 { return errors.New("File to tag and tags to apply must be specified.") }

    error := this.tagPath(args[0], args[1:])
    if error != nil { return error }

    return nil
}

// implementation

func (this TagCommand) tagPath(path string, tagNames []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    file, _, error := this.addPath(db, path)
    if error != nil { return error }

    for _, tagName := range tagNames {
        _, _, error = this.applyTag(db, path, file.Id, tagName)
        if error != nil { return error }
    }

    return nil
}

func (this TagCommand) applyTag(db *Database, path string, fileId uint, tagName string) (*Tag, *FileTag, error) {
    tag, error := db.TagByName(tagName)
    if error != nil { return nil, nil, error }

    if tag == nil {
        fmt.Printf("New tag '%v'\n", tagName)
        tag, error = db.AddTag(tagName)
        if error != nil { return nil, nil, error }
    }

    fileTag, error := db.FileTagByFileAndTag(fileId, tag.Id)
    if error != nil { return nil, nil, error }

    if fileTag == nil {
        _, error := db.AddFileTag(fileId, tag.Id)
        if error != nil { return nil, nil, error }
    } else {
        fmt.Printf("File '%v' is already tagged '%v'.\n", path, tagName)
    }

    return tag, fileTag, nil
}
