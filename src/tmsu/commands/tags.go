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

func (this TagsCommand) Summary() string {
    return "lists all tags or tags applied to a file or files"
}

func (this TagsCommand) Help() string {
    return `  tmsu tags
  tmsu tags FILE...

Without any filenames, shows the complete list of tags.

With a single filename, lists the tags applied to that file.

With multiple filenames, lists the names of these that have tags applied and the list of applied tags.`
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

                fmt.Println(path)

                for _, tag := range tags {
                    fmt.Println("  " + tag.Name)
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

    file, error := db.FileByPath(absPath)
    if error != nil { return nil, error }
    if file == nil { return nil, nil }

    tags, error := db.TagsByFileId(file.Id)
    if error != nil { return nil, error }

    return tags, error
}
