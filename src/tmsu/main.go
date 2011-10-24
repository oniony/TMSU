package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
//    debug := flag.Bool("debug", false, "Enable debugging")
	flag.Parse()

    command := flag.Arg(0)
    switch command {
        case "help": fmt.Println("Help")
        case "mount": fmt.Println("Mount")
        case "": noCommand()
        default: invalidCommand(command)
    }
}

func noCommand() {
    fmt.Fprintf(os.Stderr, "No command specified.\n")
    flag.Usage()
    os.Exit(1)
}

func invalidCommand(command string) {
    fmt.Fprintf(os.Stderr, "No such command '%v'.\n", command)
    flag.Usage()
    os.Exit(1)
}
