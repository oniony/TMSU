package main

import (
	"flag"
	"fmt"
	"os"
	"oniony.com/tmsu/db"
)

func main() {
//    debug := flag.Bool("debug", false, "Enable debugging")
	flag.Parse()

    command := flag.Arg(0)
    switch command {
        case "help": showHelp()
        case "mount": mount()
        case "tags" : tags()
        case "": missingCommand()
        default: invalidCommand(command)
    }
}

func showHelp() {
    flag.Usage()
}

func mount() {
    //TODO start fuse
}

func tags() {
    db := db.Open("/home/paul/tmsu.db")
    defer db.Close()

    tags, error := db.Tags()

    if (error != nil) {
        fmt.Fprintf(os.Stderr, "Could not retrieve tags.\nReason: %v\n", error.String())
        os.Exit(2)
    }

    for _, tag := range tags {
        fmt.Println(tag.Name)
    }
}

func missingCommand() {
    fmt.Fprintf(os.Stderr, "No command specified.\n")
    flag.Usage()
    os.Exit(1)
}

func invalidCommand(command string) {
    fmt.Fprintf(os.Stderr, "No such command '%v'.\n", command)
    flag.Usage()
    os.Exit(1)
}
