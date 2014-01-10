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
	"os"
	"tmsu/common/log"
	"tmsu/storage/database"
)

type Command struct {
	Name        string
	Synopsis    string
	Description string
	Options     Options
	Exec        func(options Options, args []string) error
}

var commands = map[string]*Command{
	"copy":    &CopyCommand,
	"delete":  &DeleteCommand,
	"dupes":   &DupesCommand,
	"files":   &FilesCommand,
	"help":    &HelpCommand,
	"merge":   &MergeCommand,
	"mount":   &MountCommand,
	"rename":  &RenameCommand,
	"repair":  &RepairCommand,
	"stats":   &StatsCommand,
	"status":  &StatusCommand,
	"tag":     &TagCommand,
	"tags":    &TagsCommand,
	"unmount": &UnmountCommand,
	"untag":   &UntagCommand,
	"version": &VersionCommand,
	"vfs":     &VfsCommand}

var globalOptions = Options{Option{"--verbose", "-v", "show verbose messages", false, ""},
	Option{"--help", "-h", "show help and exit", false, ""},
	Option{"--version", "-V", "show version information and exit", false, ""},
	Option{"--database", "-D", "use the specified database", true, ""}}

func Run() {
	helpCommands = commands

	parser := NewOptionParser(globalOptions, commands)
	commandName, options, arguments, err := parser.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case options.HasOption("--version"):
		commandName = "version"
	case options.HasOption("--help"), commandName == "":
		commandName = "help"
	}

	log.Verbosity = options.Count("--verbose") + 1

	if dbOption := options.Get("--database"); dbOption != nil && dbOption.Argument != "" {
		database.Path = dbOption.Argument
	}

	command := commands[commandName]
	if command == nil {
		log.Fatalf("Invalid command '%v'.", commandName)
	}

	err = command.Exec(options, arguments)
	if err != nil {
		log.Fatal(err.Error())
	}
}
