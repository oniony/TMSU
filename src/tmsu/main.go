package main

import (
	       "flag"
	       "fmt"
	       "log"
	       "path/filepath"
       )

var commands map [string] Command

func main() {
    commandArray := []Command{
                                 HelpCommand{},
                                 MountCommand{},
                                 UnmountCommand{},
                                 AddCommand{},
                                 TagCommand{},
                                 RenameCommand{},
                             }

    commands = make(map [string] Command, len(commandArray))
    for _, command := range commandArray { commands[command.Name()] = command }

	flag.Parse()
    if flag.NArg() == 0 { missingCommand() }
    commandName := flag.Arg(0)
    args := flag.Args()[1:]

    command := commands[commandName]
    if command == nil { log.Fatalf("No such command, '%v'.", commandName) }

    error := command.Exec(args)
    if error != nil { log.Fatal(error) }
}

// commands

func remove(args []string) {
    log.Fatal("Not implemented.")
}

func untag(args []string) {
    log.Fatal("Not implemented.")
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
