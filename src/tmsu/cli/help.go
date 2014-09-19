/*
Copyright 2011-2014 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cli

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"tmsu/cli/ansi"
	"tmsu/cli/terminal"
)

var HelpCommand = Command{
	Name:     "help",
	Synopsis: "List commands or show help for a particular subcommand",
	Description: `tmsu help [OPTION]... [SUBCOMMAND]

Shows help summary or, where SUBCOMMAND is specified, help for SUBCOMMAND.`,
	Options: Options{{"--list", "-l", "list commands", false, ""}},
	Exec:    helpExec,
}

var helpCommands map[string]*Command

func helpExec(options Options, args []string) error {
	if options.HasOption("--list") {
		listCommands()
	} else {
		switch len(args) {
		case 0:
			summary()
		default:
			commandName := args[0]
			describeCommand(commandName)
		}
	}

	return nil
}

func summary() {
	fmt.Println("TMSU")
	fmt.Println()

	var maxWidth int
	commandNames := make([]string, 0, len(helpCommands))
	for _, command := range helpCommands {
		commandName := command.Name
		maxWidth = int(math.Max(float64(maxWidth), float64(len(commandName))))
		commandNames = append(commandNames, commandName)
	}

	sort.Strings(commandNames)

	for _, commandName := range commandNames {
		command, _ := helpCommands[commandName]

		commandSummary := command.Synopsis
		if commandSummary == "" {
			continue
		}

		fmt.Printf("  %-"+strconv.Itoa(maxWidth)+"v  %v\n", command.Name, commandSummary)
	}

	fmt.Println()

	fmt.Println("Global options:")
	fmt.Println()

	for _, option := range globalOptions {
		if option.ShortName != "" && option.LongName != "" {
			fmt.Printf("  %v, %v: %v\n", option.ShortName, option.LongName, option.Description)
		} else if option.ShortName == "" {
			fmt.Printf("      %v: %v\n", option.LongName, option.Description)
		} else {
			fmt.Printf("  %v: %v\n", option.ShortName, option.Description)
		}
	}

	fmt.Println()
	fmt.Println("Specify subcommand name for detailed help on a particular subcommand")
	fmt.Println("E.g. tmsu help files")
}

func listCommands() {
	commandNames := make(ansi.Strings, 0, len(helpCommands))

	for _, command := range helpCommands {
		if command.Synopsis == "" {
			continue
		}

		commandNames = append(commandNames, ansi.String(command.Name))
	}

	sort.Sort(commandNames)

	renderColumns(commandNames, terminal.Width())
}

func describeCommand(commandName string) {
	command := helpCommands[commandName]
	if command == nil {
		fmt.Printf("No such command '%v'.\n", commandName)
		return
	}

	fmt.Println(string(command.Description))

	if len(command.Options) > 0 {
		fmt.Println()

		fmt.Println("Options:")
		fmt.Println()

		for _, option := range command.Options {
			fmt.Printf("  %v, %v: %v\n", option.ShortName, option.LongName, option.Description)
		}
	}
}
