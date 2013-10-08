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

package main

import (
	"os"
	"tmsu/cli"
	"tmsu/cli/commands"
	"tmsu/log"
)

func main() {
	helpCommand := &commands.HelpCommand{}
	commands := map[cli.CommandName]cli.Command{
		"copy":    commands.CopyCommand{},
		"delete":  commands.DeleteCommand{},
		"dupes":   commands.DupesCommand{},
		"files":   commands.FilesCommand{},
		"help":    helpCommand,
		"imply":   commands.ImplyCommand{},
		"merge":   commands.MergeCommand{},
		"mount":   commands.MountCommand{},
		"rename":  commands.RenameCommand{},
		"repair":  commands.RepairCommand{},
		"stats":   commands.StatsCommand{},
		"status":  commands.StatusCommand{},
		"tag":     commands.TagCommand{},
		"tags":    commands.TagsCommand{},
		"unmount": commands.UnmountCommand{},
		"untag":   commands.UntagCommand{},
		"version": commands.VersionCommand{},
		"vfs":     commands.VfsCommand{},
	}
	helpCommand.Commands = commands

	globalOptions := cli.Options{cli.Option{"--verbose", "-v", "show verbose messages", false, ""},
		cli.Option{"--help", "-h", "show help and exit", false, ""},
		cli.Option{"--version", "-V", "show version information and exit", false, ""}}

	parser := cli.NewOptionParser(globalOptions, commands)
	commandName, options, arguments, err := parser.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if commandName == "" {
		commandName = "help"
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
