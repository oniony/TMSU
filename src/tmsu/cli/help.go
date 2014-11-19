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
	"tmsu/common/terminal"
	"tmsu/common/terminal/ansi"
)

var HelpCommand = Command{
	Name:     "help",
	Synopsis: "List subcommands or show help for a particular subcommand",
	Description: `$BOLDtmsu help [OPTION]... [SUBCOMMAND]$RESET

Shows help summary or, where SUBCOMMAND is specified, help for SUBCOMMAND.`,
	Options: Options{{"--list", "-l", "list commands", false, ""}},
	Exec:    helpExec,
}

var helpCommands map[string]*Command

func helpExec(options Options, args []string) error {
	var colour bool
	if options.HasOption("--color") {
		when := options.Get("--color").Argument
		switch when {
		case "auto":
			colour = terminal.Colour() && terminal.Width() > 0
		case "":
		case "always":
			colour = true
		case "never":
			colour = false
		default:
			return fmt.Errorf("invalid argument '%v' for '--color'", when)
		}
	} else {
		colour = terminal.Colour() && terminal.Width() > 0
	}

	if options.HasOption("--list") {
		listCommands()
	} else {
		switch len(args) {
		case 0:
			summary()
		default:
			commandName := args[0]
			describeCommand(commandName, colour)
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

	terminal.PrintList(commandNames, terminal.Width(), false)
}

func describeCommand(commandName string, colour bool) {
	command := findCommand(helpCommands, commandName)
	if command == nil {
		fmt.Printf("No such command '%v'.\n", commandName)
		return
	}

	description := ansi.ParseMarkup(command.Description)

	terminal.PrintWrapped(ansi.String(description), terminal.Width(), colour)

	if command.Aliases != nil && len(command.Aliases) > 0 {
		fmt.Println()

		if command.Aliases != nil {
			terminal.Print(ansi.Bold+ansi.String("Aliases:")+ansi.Reset, colour)

			for _, alias := range command.Aliases {
				fmt.Print(" " + alias)
			}
		}

		fmt.Println()
	}

	if len(command.Options) > 0 {
		fmt.Println()

		terminal.Println(ansi.Bold+ansi.String("Options:")+ansi.Reset, colour)
		fmt.Println()

		for _, option := range command.Options {
			fmt.Printf("  %v, %v: %v\n", option.ShortName, option.LongName, option.Description)
		}
	}
}
