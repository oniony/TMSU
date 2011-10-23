package main

import (
	"fmt"
	"flag"
)

type Command interface {
	Execute()
}

func main() {
	flag.Parse()

	commands := map[string] Command {
		"help" : new(HelpCommand),
	}

	for _, command := range commands {
		command.Execute()
	}

	for i := 0; i < flag.NArg(); i += 1 {
		fmt.Println(flag.Args()[i])
	}
}
