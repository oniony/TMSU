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

func (this HelpCommand) Description() string {
    return "lists commands or provides help for a particular command"
}

func (this HelpCommand) Exec(args []string) error {
    fmt.Println("tmsu")
    fmt.Println()

    var maxWidth uint = 0
    for _, command := range commands {
        maxWidth = uint(math.Max(float64(maxWidth), float64(len(command.Name()))))
    }

    for _, command := range commands {
        fmt.Printf("  %-" + strconv.Uitoa(maxWidth) + "v  %v\n", command.Name(), command.Description())
    }

    fmt.Println()
    fmt.Println("Copyright (C) 2011 Paul Ruane")

    return nil
}
