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
	globalOptions []GlobalOption
}

type GlobalOption string

type CommandName string

type CommandOption string

type Argument string

func Create() *Parser {
	parser := Parser{[]GlobalOption{"-d", "--debug"}}
	return &parser
}

func (parser *Parser) Parse(args []string) ([]GlobalOption, CommandName, []CommandOption, []Argument, error) {
	globalOptions := make([]GlobalOption, 0)
	commandName := CommandName("")
	commandOptions := make([]CommandOption, 0)
	arguments := make([]Argument, 0)

	index := 0
	for ; index < len(args); index += 1 {
		arg := args[index]

		if arg[0] == '-' {
			globalOption := GlobalOption(arg)
			if parser.globalOptionExists(globalOption) {
				globalOptions = append(globalOptions, globalOption)
			} else {
				return globalOptions, commandName, commandOptions, arguments, errors.New("Invalid option '" + arg + "'.")
			}
		} else {
			commandName = CommandName(arg)
			break
		}
	}

	//TODO lookup command
	//TODO error if invalid command

	for index += 1; index < len(args); index += 1 {
		arg := args[index]

		if arg[0] == '-' {
			commandOption := CommandOption(arg)
			//TODO check if command supports arg
			commandOptions = append(commandOptions, commandOption)
		} else {
			arguments = append(arguments, Argument(arg))
		}
	}

	return globalOptions, commandName, commandOptions, arguments, nil
}

func (parser *Parser) globalOptionExists(option GlobalOption) bool {
	for _, globalOption := range parser.globalOptions {
		if globalOption == option {
			return true
		}
	}

	return false
}
