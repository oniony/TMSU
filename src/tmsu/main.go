package main

import (
	"flag"
	"fmt"
	"os"
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
        case "tag": tag(args)
        case "untag": untag(args)
        case "tags" : tags(args)
        default: invalidCommand(command)
    }
}

// commands

func help() {
    showUsage()
    os.Exit(0)
}

func mount(args []string) {
    if (len(args) == 0) { die("No mountpoint specified.") }

    mountPath := args[0]

    vfs, error := MountVfs(mountPath)
    if (error != nil) { die("Could not mount filesystem: %v", error.String()) }
    defer vfs.Unmount()

    fmt.Printf("Database '%v' mounted at '%v'.\n", DatabasePath(), mountPath)

    vfs.Loop()
}

func add(args []string) {
    fmt.Println("Not implemented.")
}

func remove(args []string) {
    fmt.Println("Not implemented.")
}

func tag(args []string) {
    fmt.Println("Not implemented.")
}

func untag(args []string) {
    fmt.Println("Not implemented.")
}

func tags(args []string) {
    db, error := OpenDatabase(DatabasePath())
    if error != nil { die("Could not open database: %v", error.String()) }
    defer db.Close()

    tags, error := db.Tags()
    if error != nil { die("Could not retrieve tags: %v", error.String()) }

    for _, tag := range tags {
        fmt.Println(tag.Name)
    }
}

// other stuff

func showUsage() {
    fmt.Println("usage: tmsu <command> [<args>]")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println()
    fmt.Println(" help       list commands or provide help for a given command")
    fmt.Println(" add        add a file to the VFS without applying tags")
    fmt.Println(" tag        add a file (if necessary) and apply tags")
    fmt.Println(" tags       list all tags or tags for a given file")
}

func missingCommand() {
    die("No command specified.")
}

func invalidCommand(command string) {
    die("No such command '%v'.", command)
}
