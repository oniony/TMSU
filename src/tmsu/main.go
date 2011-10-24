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
        case "": missingCommand()
        default: invalidCommand(command)
    }
}

func showHelp() {
    flag.Usage()
}

func mount() {
    db := db.Open("/home/paul/tmsu.db")
    defer db.Close() 

    fmt.Println("DB: ", db)
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
