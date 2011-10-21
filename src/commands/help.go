package commands

import (
	"fmt"
)

type HelpCommand struct {}

func (HelpCommand) Execute() {
	fmt.Println("Help!")
}
