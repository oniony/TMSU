package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()

    if flag.NArg() == 0 { missingCommand() }

    command := flag.Arg(0)
    args := flag.Args()[1:]
    if command == "help" { help() }

    switch command {
        case "mount": mount(args)

        case "add": add(args)
        case "remove": remove(args)
        case "dupes": dupes(args)

        case "tag": tag(args)
        case "untag": untag(args)
        case "tags" : tags(args)
        case "rename": rename(args)

        default: invalidCommand(command)
    }
}

// commands

func help() {
    showUsage()
    os.Exit(0)
}

func mount(args []string) {
    if (len(args) == 0) { log.Fatal("No mountpoint specified.") }

    mountPath := args[0]

    vfs, error := MountVfs(mountPath)
    if (error != nil) { log.Fatalf("Could not mount filesystem: %v", error) }
    defer vfs.Unmount()

    log.Printf("Database '%v' mounted at '%v'.\n", databasePath(), mountPath)

    vfs.Loop()
}

func add(paths []string) {
    db, err := OpenDatabase(databasePath())
    if err != nil { log.Fatalf("Could not open database: %v", err) }
    defer db.Close()

    for _, path := range paths {
        addPath(db, path)
    }
}

func remove(args []string) {
    log.Fatal("Not implemented.")
}

func tag(args []string) {
    path := args[0]
    tagNames := args[1:]

    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    //TODO check args length

    file, _, error := addPath(db, path)
    if error != nil { log.Fatalf("Could not add file: %v", error) }

    for _, tagName := range tagNames {
        tagPath(db, path, file.Id, tagName)
    }
}

func untag(args []string) {
    log.Fatal("Not implemented.")
}

func rename(args []string) {
    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    currentName := args[0]
    newName := args[1]

    tag, error := db.TagByName(currentName)
    if error != nil { log.Fatalf("Could not retrieve tag '%v': %v", currentName, error) }
    if tag == nil { log.Fatalf("No such tag '%v'.", currentName) }

    newTag, error := db.TagByName(newName)
    if error != nil { log.Fatalf("Could not retrieve tag '%v': %v", newName, error) }

    //TODO only merge if flag set

    if newTag == nil {
        _, error = db.RenameTag(tag.Id, newName)
        if error != nil { log.Fatalf("Could not rename tag '%v': %v", currentName, error) }
    } else {
        error = db.MigrateFileTags(tag.Id, newTag.Id)
        if error != nil { log.Fatalf("Could not merge tags '%v' into '%v': %v", currentName, newName, error) }

        error = db.DeleteTag(tag.Id)
        if error != nil { log.Fatalf("Tag '%v' merged to '%v' but old tag could not be deleted: %v", currentName, newName, error) }
    }
}

func tags(args []string) {
    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    switch len(args) {
        case 0: listAllTags(db)
        case 1: listTagsForPath(db, args[0])
        default:
            for _, path := range args {
                fmt.Println(path)
                listTagsForPath(db, path)
            }
    }
}

func dupes(args []string) {
    log.Fatal("Not implemented.")
}

// other stuff

func showUsage() {
    fmt.Println("usage: tmsu <command> [<args>]")
    fmt.Println()
    fmt.Println("commands:")
    fmt.Println()
    fmt.Println(" help       list commands or provide help for a given command")
    fmt.Println(" mount      mounts the file-system")
    fmt.Println(" add        add a file without applying tags")
    fmt.Println(" remove     remove a file")
    fmt.Println(" tag        add a file (if necessary) and apply tags")
    fmt.Println(" untag      disassociate a file with tags")
    fmt.Println(" tags       list all tags or tags for a given file")
    fmt.Println(" dupes      list duplicate files")
}

func missingCommand() {
    log.Fatal("No command specified.")
}

func invalidCommand(command string) {
    log.Fatalf("No such command '%v'.", command)
}

func addPath(db *Database, path string) (*File, *FilePath, error) {
    absPath, error := filepath.Abs(path)
    if error != nil { log.Fatalf("Could resolve path '%v': %v", path, error) }

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
            fmt.Printf("Adding file '%v' (duplicate of a previously added file).\n", path)

            filePath, error = db.AddFilePath(file.Id, absPath)
            if error != nil { return nil, nil, error }
        case uniqueContents:
            fmt.Printf("Updating with modified file '%v'.\n", path)

            //TODO contents have changed, update file-path
            log.Fatalf("Not implemented.")
        default:
            fmt.Printf("File '%v' is already added.\n", path)
    }

    return file, filePath, nil
}

func tagPath(db *Database, path string, fileId uint, tagName string) (*Tag, *FileTag, error) {
    tag, error := db.TagByName(tagName)
    if error != nil { return nil, nil, error }

    if tag == nil {
        tag, error = db.AddTag(tagName)
        if error != nil { return nil, nil, error }
    }

    fileTag, error := db.FileTagByFileAndTag(fileId, tag.Id)
    if error != nil { return nil, nil, error }

    if fileTag == nil {
        fmt.Printf("Tagged file '%v' with '%v'.\n", path, tagName)

        _, error := db.AddFileTag(fileId, tag.Id)
        if error != nil { return nil, nil, error }
    } else {
        fmt.Printf("File '%v' is already tagged '%v'.\n", path, tagName)
    }

    return tag, fileTag, nil
}

func listAllTags(db *Database) {
    tags, error := db.Tags()
    if error != nil { log.Fatalf("Could not retrieve tags: %v", error) }

    for _, tag := range tags {
        fmt.Println(tag.Name)
    }
}

func listTagsForPath(db *Database, path string) error {
    absPath, error := filepath.Abs(path)
    if error != nil { log.Fatalf("Could resolve path '%v': %v", path, error) }

    fingerprint, error := Fingerprint(absPath)
    if error != nil { return error }

    file, error := db.FileByFingerprint(fingerprint)
    if error != nil { return error }
    if file == nil { return nil }

    tags, error := db.TagsByFile(file.Id)
    if error != nil { return error }

    for _, tag := range tags {
        fmt.Printf("  %v\n", tag.Name)
    }

    return nil
}
