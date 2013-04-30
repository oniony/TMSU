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

type Command interface {
	Name() CommandName
	Synopsis() string
	Description() string
	Options() Options
	Exec(options Options, args []string) error
}

type CommandName string

type CommandNames []CommandName

func (commandNames CommandNames) Len() int {
	return len(commandNames)
}

func (commandNames CommandNames) Less(i, j int) bool {
	return commandNames[i] < commandNames[j]
}

func (commandNames CommandNames) Swap(i, j int) {
	commandNames[i], commandNames[j] = commandNames[j], commandNames[i]
}

var globalOptions = Options{Option{"--verbose", "-v", "show verbose messages", false, ""},
	Option{"--help", "-h", "show help and exit", false, ""},
	Option{"--version", "-V", "show version information and exit", false, ""}}
