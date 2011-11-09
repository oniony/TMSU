package main

import (
           "fmt"
       )

type HelpCommand struct {}

func (this HelpCommand) Name() string {
    return "help"
}

func (this HelpCommand) Description() string {
    return "lists commands or provides help for a particular command"
}

func (this HelpCommand) Exec(args []string) error {
    fmt.Println("TMSU")
    fmt.Println()

    //TODO work out max width of command names and use in formatting
    for _, command := range commands {
        fmt.Printf("    %10v    %v\n", command.Name(), command.Description())
    }

    return nil
}
