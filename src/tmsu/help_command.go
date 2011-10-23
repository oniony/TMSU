package main

import (
	"fmt"
)

type HelpCommand struct {}

func (HelpCommand) Execute() {
	fmt.Println("TMSU")
	fmt.Println();
}
