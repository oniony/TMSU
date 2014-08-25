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

var globalOptions = Options{Option{"--verbose", "-v", "show verbose messages", false, ""},
	Option{"--help", "-h", "show help and exit", false, ""},
	Option{"--version", "-V", "show version information and exit", false, ""},
	Option{"--database", "-D", "use the specified database", true, ""},
	Option{"--color", "", "use color to show implied tags (auto/always/never)", true, ""},
}

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
		log.Fatalf("invalid command '%v'.", commandName)
	}

	err = command.Exec(options, arguments)
	if err != nil {
		if err != errBlank {
			log.Warn(err.Error())
		}

		os.Exit(1)
	}
}
