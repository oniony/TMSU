// Copyright 2011 Paul Ruane. All rights reserved.

package main

import (
	"fmt"
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

	var commandName string
	if len(os.Args) > 1 {
		commandName = os.Args[1]
	} else {
		commandName = "help"
	}

	command := commands[commandName]
	if command == nil {
		fmt.Printf("No such command, '%v'.\n", commandName)
		os.Exit(1)
	}

	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	} else {
		args = []string{}
	}

	error := command.Exec(args)
	if error != nil {
	    fmt.Fprintln(os.Stderr, error.Error())
	    os.Exit(1)
	}
}
