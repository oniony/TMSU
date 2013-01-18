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

type Option struct {
	LongName    string
	ShortName   string
	Description string
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
