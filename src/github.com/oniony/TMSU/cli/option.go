// Copyright 2011-2017 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"strings"
)

type Option struct {
	LongName    string
	ShortName   string
	Description string
	HasArgument bool
	Argument    string
}

type Options []Option

func (options Options) HasOption(name string) bool {
	for _, option := range options {
		if option.LongName == name || option.ShortName == name {
			return true
		}
	}

	return false
}

func (options Options) Count(name string) uint {
	var count uint

	for _, option := range options {
		if option.LongName == name || option.ShortName == name {
			count++
		}
	}

	return count
}

func (options Options) Get(name string) *Option {
	for _, option := range options {
		if option.LongName == name || option.ShortName == name {
			return &option
		}
	}

	return nil
}

type OptionParser struct {
	globalOptions Options
	commandByName map[string]*Command
}

func NewOptionParser(globalOptions Options, commands []*Command) *OptionParser {
	parser := OptionParser{globalOptions, buildCommandByNameMap(commands)}
	return &parser
}

func (parser *OptionParser) Parse(args ...string) (command *Command, options Options, arguments []string, err error) {
	commandName := ""
	options = make(Options, 0)
	arguments = make([]string, 0)

	possibleOptions := make(Options, len(globalOptions))
	copy(possibleOptions, globalOptions)

	parseOptions := true
	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch {
		case arg == "":
			err = fmt.Errorf("invalid empty argument")
			return
		case arg == "--" && parseOptions:
			parseOptions = false
		default:
			if parseOptions && len(arg) > 1 && arg[0] == '-' {
				parts := strings.SplitN(arg, "=", 2)
				optionName := parts[0]

				option := lookupOption(possibleOptions, optionName)
				if option == nil {
					err = fmt.Errorf("invalid option '%v'", optionName)
					return
				}

				if option.HasArgument {
					if len(parts) == 2 {
						option.Argument = parts[1]
					} else {
						if len(args) < index+2 {
							err = fmt.Errorf("missing argument for option '%v'", optionName)
							return
						}

						option.Argument = args[index+1]
						index++
					}
				}

				options = append(options, *option)
			} else {
				if commandName == "" {
					commandName = arg

					var ok bool
					command, ok = parser.commandByName[commandName]
					if ok {
						possibleOptions = append(possibleOptions, command.Options...)
					} else {
						err = fmt.Errorf("invalid subcommand '%v'", commandName)
					}
				} else {
					arguments = append(arguments, arg)
				}
			}
		}
	}

	return
}

// unexported

func buildCommandByNameMap(commands []*Command) map[string]*Command {
	commandByName := make(map[string]*Command)

	for _, command := range commands {
		commandByName[command.Name] = command

		for _, alias := range command.Aliases {
			commandByName[alias] = command
		}
	}

	return commandByName
}

func lookupOption(options Options, name string) *Option {
	for _, option := range options {
		if option.ShortName == name || option.LongName == name {
			return &Option{option.LongName, option.ShortName, option.Description, option.HasArgument, ""}
		}
	}

	return nil
}
