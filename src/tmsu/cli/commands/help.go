/*
Copyright 2011-2013 Paul Ruane.

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

package commands

import (
	"math"
	"sort"
	"strconv"
	"tmsu/cli"
	"tmsu/log"
)

type HelpCommand struct {
	Commands map[cli.CommandName]cli.Command
}

func (HelpCommand) Name() cli.CommandName {
	return "help"
}

func (HelpCommand) Synopsis() string {
	return "List commands or show help for a particular command"
}

func (HelpCommand) Description() string {
	return `tmsu help [OPTION]... [COMMAND]

Shows help summary or, where COMMAND is specified, help for COMMAND.`
}

func (HelpCommand) Options() cli.Options {
	return cli.Options{{"--list", "-l", "list commands", false, ""}}
}

func (command HelpCommand) Exec(options cli.Options, args []string) error {
	if options.HasOption("--list") {
		command.listCommands()
	} else {
		switch len(args) {
		case 0:
			command.summary()
		default:
			commandName := cli.CommandName(args[0])
			command.describeCommand(commandName)
		}
	}

	return nil
}

func (helpCommand HelpCommand) summary() {
	log.Print("TMSU")
	log.Print()

	var maxWidth int = 0
	commandNames := make(cli.CommandNames, 0, len(helpCommand.Commands))
	for _, command := range helpCommand.Commands {
		commandName := command.Name()
		maxWidth = int(math.Max(float64(maxWidth), float64(len(commandName))))
		commandNames = append(commandNames, commandName)
	}

	sort.Sort(commandNames)

	for _, commandName := range commandNames {
		command, _ := helpCommand.Commands[commandName]

		commandSummary := command.Synopsis()
		if commandSummary == "" {
			continue
		}

		log.Printf("  %-"+strconv.Itoa(maxWidth)+"v  %v", command.Name(), commandSummary)
	}
}

func (helpCommand HelpCommand) listCommands() {
	commandNames := make(cli.CommandNames, 0, len(helpCommand.Commands))

	for _, command := range helpCommand.Commands {
		if command.Synopsis() == "" {
			continue
		}

		commandNames = append(commandNames, command.Name())
	}

	sort.Sort(commandNames)

	for _, commandName := range commandNames {
		log.Print(commandName)
	}
}

func (helpCommand HelpCommand) describeCommand(commandName cli.CommandName) {
	command := helpCommand.Commands[commandName]
	if command == nil {
		log.Printf("No such command '%v'.", commandName)
		return
	}

	log.Print(command.Description())

	if len(command.Options()) > 0 {
		log.Print()

		log.Print("Options:")
		log.Print()

		for _, option := range command.Options() {
			log.Printf("  %v, %v: %v", option.ShortName, option.LongName, option.Description)
		}
	}

	log.Print()
}
