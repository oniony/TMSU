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

package cli

import (
	"fmt"
)

type Parser struct {
	commandByName map[CommandName]Command
}

func NewParser(commandByName map[CommandName]Command) *Parser {
	parser := Parser{commandByName}
	return &parser
}

func (parser *Parser) Parse(args []string) (CommandName, Options, []string, error) {
	optionNames := make([]string, 0)
	arguments := make([]string, 0)

	parseOptions := true
	for _, arg := range args {
		if arg == "--" {
			parseOptions = false
		} else {
			if parseOptions && arg[0] == '-' {
				optionNames = append(optionNames, arg)
			} else {
				arguments = append(arguments, arg)
			}
		}
	}

	commandName := CommandName("")

	if len(arguments) > 0 {
		commandName = CommandName(arguments[0])
		arguments = arguments[1:]
	}

	command := parser.commandByName[commandName]

	options := make(Options, 0)
	for _, optionName := range optionNames {
		option := LookupOption(command, optionName)

		if option == nil {
			return "", nil, nil, fmt.Errorf("invalid option '%v'.", optionName)
		}

		options = append(options, *option)
	}

	if options.HasOption("--version") {
		commandName = "version"
	}
	if options.HasOption("--help") {
		commandName = "help"
	}

	return commandName, options, arguments, nil
}
