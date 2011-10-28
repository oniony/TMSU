package main

import (
	"flag"
	"fmt"
	"os"
	"oniony.com/tmsu/db"
	"oniony.com/tmsu/vfs"
)

func main() {
	flag.Parse()

    checkDatabase()

    command := flag.Arg(0)
    switch command {
        case "help": showUsage()
        case "mount": mount(flag.Args())
        case "tags" : tags(flag.Args())
        case "": missingCommand()
        default: invalidCommand(command)
    }
}

func die(format string, a ...interface{}) {
    fmt.Fprintf(os.Stderr, format + "\n", a...)
    os.Exit(1)
}

func showUsage() {
    fmt.Println("usage: tmsu <command> [<args>]")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println()
    fmt.Println(" help       list commands or provide help for a given command")
    fmt.Println(" add        add a file to the VFS without applying tags")
    fmt.Println(" tag        add a file (if necessary) and apply tags")
    fmt.Println(" tags       list all tags or tags for a given file")
    os.Exit(0)
}

func checkDatabase() {
    databasePath := DatabasePath()
    _, error := os.Open(databasePath)

    if (error == nil) { return }

    switch error.(type) {
        case *os.PathError:
            switch error.(*os.PathError).Error {
                case os.ENOENT: die("No database at '%v'.", databasePath)
                default: die("(PathError) Could not open database '%v': %v", databasePath, error.(*os.PathError))
            }
        default: die("Could not open database '%v': %v", databasePath, error)
    }
}

func mount(args []string) {
    vfs, error := vfs.Mount("./mountpoint")
    if (error != nil) { die("Could not mount filesystem: %v", error.String()) }

    vfs.Loop()
}

func tags(args []string) {
    db := db.Open(DatabasePath())
    defer db.Close()

    tags, error := db.Tags()
    if (error != nil) { die("Could not retrieve tags: %v", error.String()) }

    for _, tag := range tags {
        fmt.Println(tag.Name)
    }
}

func missingCommand() {
    die("No command specified.")
}

func invalidCommand(command string) {
    die("No such command '%v'.", command)
}
