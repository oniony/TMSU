package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var commands map[string]Command

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
		VfsCommand{},
	}

	commands = make(map[string]Command, len(commandArray))
	for _, command := range commandArray {
		commands[command.Name()] = command
	}

	flag.Parse()

	var commandName string
	if flag.NArg() > 0 {
		commandName = flag.Arg(0)
	} else {
		commandName = "help"
	}

	command := commands[commandName]
	if command == nil {
		fmt.Printf("No such command, '%v'.\n", commandName)
		os.Exit(1)
	}

	var args []string
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	} else {
		args = []string{}
	}

	error := command.Exec(args)
	if error != nil {
		log.Fatal(error)
	}
}
