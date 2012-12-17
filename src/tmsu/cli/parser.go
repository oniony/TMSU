/*
Copyright 2011-2012 Paul Ruane.

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
	"errors"
)

type Parser struct {
	commandByName map[CommandName]Command
	globalOptions Options
}

func NewParser(commandByName map[CommandName]Command) *Parser {
	parser := Parser{commandByName,
		Options{Option{"-v", "--verbose", "show verbose messages"},
			Option{"-h", "--help", "show help and exit"},
			Option{"-V", "--version", "show version information and exit"}}}
	return &parser
}

func (parser *Parser) Parse(args []string) (Options, CommandName, Options, []string, error) {
	globalOptions := make(Options, 0)
	commandName := CommandName("help")
	commandOptions := make(Options, 0)
	arguments := make([]string, 0)

	index := 0
	for ; index < len(args); index += 1 {
		arg := args[index]

		if arg[0] == '-' {
			globalOption := parser.lookupGlobalOption(arg)
			if globalOption != nil {
				switch globalOption.LongName {
				case "--version":
					return Options{}, "version", Options{}, []string{}, nil
				case "--help":
					return Options{}, "help", Options{}, []string{}, nil
				default:
					globalOptions = append(globalOptions, *globalOption)
				}
			} else {
				return globalOptions, commandName, commandOptions, arguments, errors.New("Invalid global option '" + arg + "'.")
			}
		} else {
			commandName = CommandName(arg)
			break
		}
	}

	command := parser.commandByName[commandName]

	if command != nil {
		for index += 1; index < len(args); index += 1 {
			arg := args[index]

			if arg[0] == '-' {
				commandOption := parser.lookupCommandOption(command, arg)
				if commandOption == nil {
					return globalOptions, commandName, commandOptions, arguments, errors.New("Invalid command option '" + arg + "'.")
				}

				commandOptions = append(commandOptions, *commandOption)
			} else {
				arguments = append(arguments, arg)
			}
		}
	}

	return globalOptions, commandName, commandOptions, arguments, nil
}

func (parser *Parser) lookupGlobalOption(option string) *Option {
	for _, globalOption := range parser.globalOptions {
		if globalOption.ShortName == option || globalOption.LongName == option {
			return &globalOption
		}
	}

	return nil
}

func (parser *Parser) lookupCommandOption(command Command, option string) *Option {
	for _, commandOption := range command.Options() {
		if commandOption.ShortName == option || commandOption.LongName == option {
			return &commandOption
		}
	}

	return nil
}
