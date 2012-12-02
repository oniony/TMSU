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

package main

import (
	"os"
	"tmsu/cli"
	"tmsu/cli/commands"
	"tmsu/common"
)

func main() {
	helpCommand := &commands.HelpCommand{}
	commands := map[string]cli.Command{
		"copy":    commands.CopyCommand{},
		"delete":  commands.DeleteCommand{},
		"dupes":   commands.DupesCommand{},
		"files":   commands.FilesCommand{},
		"help":    helpCommand,
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

	aliases := map[string]string{
		"-?":        "help",
		"-h":        "help",
		"-help":     "help",
		"--help":    "help",
		"-V":        "version",
		"-version":  "version",
		"--version": "version",
	}

	args := os.Args[1:] // strip off binary name

	var commandName string
	if len(args) > 0 {
		commandName = args[0]
		args = args[1:]
	} else {
		commandName = "help"
	}

	dealiased, found := aliases[commandName]
	if found {
		commandName = dealiased
	}

	command := commands[commandName]
	if command == nil {
		common.Fatalf("unknown command '%v'.", commandName)
	}

	err := command.Exec(args)
	if err != nil {
		common.Fatal(err.Error())
	}
}
