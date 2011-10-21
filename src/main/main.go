package main

import (
	"fmt"
	"flag"
	"./obj/commands"
)

type Command interface {
	Execute()
}

func main() {
	flag.Parse()

	commands := map[string] Command {
		"help" : new(commands.HelpCommand),
	}

	for name, command := range commands {
		fmt.Println("Command: ", name)

		command.Execute()
	}

	for i := 0; i < flag.NArg(); i += 1 {
		fmt.Println(flag.Args()[i])
	}
}
