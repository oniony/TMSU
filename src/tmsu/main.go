package main

import (
	       "flag"
	       "log"
       )

var commands map [string] Command

func main() {
    commandArray := []Command{
                                 HelpCommand{},
                                 MountCommand{},
                                 UnmountCommand{},
                                 TagsCommand{},
                                 TagCommand{},
                                 UntagCommand{},
                                 RenameCommand{},
                                 MergeCommand{},
                                 DeleteCommand{},
                                 ExportCommand{},
                             }

    commands = make(map [string] Command, len(commandArray))
    for _, command := range commandArray { commands[command.Name()] = command }

	flag.Parse()

	var commandName string
    if flag.NArg() > 0 {
        commandName = flag.Arg(0)
    } else {
        commandName = "help"
    }

    var args []string
    if flag.NArg() > 1 {
        args = flag.Args()[1:]
    } else {
        args = []string {}
    }

    command := commands[commandName]
    if command == nil { log.Fatalf("No such command, '%v'.", commandName) }

    error := command.Exec(args)
    if error != nil { log.Fatal(error) }
}
