/*
Copyright 2011 Paul Ruane.

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

var Commands map[string]Command

func init() {
	commandArray := []Command{
		DeleteCommand{},
		DupesCommand{},
		ExportCommand{},
		FilesCommand{},
		HelpCommand{},
		MergeCommand{},
		MountCommand{},
		RenameCommand{},
		StatsCommand{},
		StatusCommand{},
		TagCommand{},
		TagsCommand{},
		UnmountCommand{},
		UntagCommand{},
		VersionCommand{},
		VfsCommand{},
	}

	Commands = make(map[string]Command, len(commandArray))
	for _, command := range commandArray {
		Commands[command.Name()] = command
	}
}

type Command interface {
	Name() string
	Synopsis() string
	Description() string
	Exec(args []string) error
}
