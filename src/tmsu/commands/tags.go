package main

import (
           "fmt"
           "log"
           "path/filepath"
       )

type TagsCommand struct {}

func (this TagsCommand) Name() string {
    return "tags"
}

func (this TagsCommand) Description() string {
    return "lists all tags or tags applied to a file or files"
}

func (this TagsCommand) Exec(args []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    switch len(args) {
        case 0:
            tags, error := this.allTags(db)
            if error != nil { log.Fatalf("Could not retrieve tags: %v", error) }

            for _, tag := range tags {
                fmt.Println(tag.Name)
            }
        case 1:
            path := args[0]

            tags, error := this.tagsForPath(db, path)
            if error != nil { log.Fatalf("Could not retrieve tags for '%v': %v", path, error) }

            for _, tag := range tags {
                fmt.Println(tag.Name)
            }
        default:
            for _, path := range args {
                tags, error := this.tagsForPath(db, path)
                if error != nil { log.Fatalf("Could not retrieve tags for '%v': %v", path, error) }

                if len(tags) > 0 {
                    fmt.Println(path)

                    for _, tag := range tags {
                        fmt.Println("  " + tag.Name)
                    }
                }
            }
    }

    return nil
}

// implementation

func (this TagsCommand) allTags(db *Database) ([]Tag, error) {
    tags, error := db.Tags()
    if error != nil { return nil, error }

    return tags, nil
}

func (this TagsCommand) tagsForPath(db *Database, path string) ([]Tag, error) {
    absPath, error := filepath.Abs(path)
    if error != nil { return nil, error }

    filePath, error := db.FilePathByPath(absPath)
    if error != nil { return nil, error }

    tags, error := db.TagsByFile(filePath.FileId)
    if error != nil { return nil, error }

    return tags, error
}
