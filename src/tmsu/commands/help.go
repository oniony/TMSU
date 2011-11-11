package main

import (
           "fmt"
           "math"
           "strconv"
       )

type HelpCommand struct {}

func (this HelpCommand) Name() string {
    return "help"
}

func (this HelpCommand) Summary() string {
    return "lists commands or provides help for a particular command"
}

func (this HelpCommand) Help() string {
    return `  tmsu help          lists commands
  tmsu help COMMAND  shows help for COMMAND

Shows command summary or, when a command is specified, detailed help for that command.`
}

func (this HelpCommand) Exec(args []string) error {
    switch len(args) {
        case 0:
            this.overview()
        default:
            this.commandHelp(args[0])
    }

    return nil
}

// implementation

func (this HelpCommand) overview() {
    fmt.Println("tmsu v0.1 (alpha)")
    fmt.Println()

    var maxWidth uint = 0
    for _, command := range commands {
        maxWidth = uint(math.Max(float64(maxWidth), float64(len(command.Name()))))
    }

    for _, command := range commands {
        fmt.Printf("  %-" + strconv.Uitoa(maxWidth) + "v  %v\n", command.Name(), command.Summary())
    }

    fmt.Println()
    fmt.Println("Copyright (C) 2011 Paul Ruane")
}

func (this HelpCommand) commandHelp(commandName string) {
    command := commands[commandName]
    if command == nil {
        fmt.Printf("No such command '%v'.\n", commandName)
        return
    }

    fmt.Println(command.Help())
}
