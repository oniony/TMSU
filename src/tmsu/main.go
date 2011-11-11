package main

import (
	       "flag"
	       "fmt"
	       "log"
       )

var commands map [string] Command

func main() {
    commandArray := []Command{
                                 HelpCommand{},
                                 MountCommand{},
                                 UnmountCommand{},
                                 AddCommand{},
                                 RemoveCommand{},
                                 TagsCommand{},
                                 TagCommand{},
                                 UntagCommand{},
                                 RenameCommand{},
                             }

    commands = make(map [string] Command, len(commandArray))
    for _, command := range commandArray { commands[command.Name()] = command }

	flag.Parse()
    if flag.NArg() == 0 { showUsage() }
    commandName := flag.Arg(0)
    args := flag.Args()[1:]

    command := commands[commandName]
    if command == nil { log.Fatalf("No such command, '%v'.", commandName) }

    error := command.Exec(args)
    if error != nil { log.Fatal(error) }
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

