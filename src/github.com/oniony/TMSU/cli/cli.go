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
	"github.com/oniony/TMSU/common/log"
	_path "github.com/oniony/TMSU/common/path"
	"os"
	"os/user"
	"path/filepath"
)

func Run() {
	helpCommands = commands

	parser := NewOptionParser(globalOptions, commands)
	command, options, arguments, err := parser.Parse(os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case options.HasOption("--version"):
		command = findCommand(commands, "version")
	case options.HasOption("--help"), command == nil:
		command = findCommand(commands, "help")
	}

	log.Verbosity = options.Count("--verbose") + 1

	var databasePath string
	switch {
	case options.HasOption("--database"):
		log.Infof(2, "using database from command-line option")
		databasePath = options.Get("--database").Argument
	case os.Getenv("TMSU_DB") != "":
		log.Infof(2, "using database from environment variable")
		databasePath = os.Getenv("TMSU_DB")
	default:
		databasePath, err = findDatabase()
		if err != nil {
			log.Fatalf("could not find database: %v", err)
		}
	}

	err, warnings := command.Exec(options, arguments, databasePath)

	if warnings != nil {
		for _, warning := range warnings {
			log.Warn(warning)
		}
	}

	if err != nil {
		log.Warn(err.Error())
	}

	if err != nil || (warnings != nil && len(warnings) > 0) {
		os.Exit(1)
	}
}

// unexported

var globalOptions = Options{Option{"--verbose", "-v", "show verbose messages", false, ""},
	Option{"--help", "-h", "show help and exit", false, ""},
	Option{"--version", "-V", "show version information and exit", false, ""},
	Option{"--database", "-D", "use the specified database", true, ""},
	Option{"--color", "", "colorize the output (auto/always/never)", true, ""},
}

func findDatabase() (string, error) {
	databasePath, err := findDatabaseInPath()
	if err != nil {
		return "", err
	}
	if databasePath != "" {
		return databasePath, nil
	}

	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("could not identify current user: %v", err))
	}

	return filepath.Join(u.HomeDir, ".tmsu", "default.db"), nil
}

func findDatabaseInPath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// look for .tmsu/db in current directory and ancestors
	for {
		dbPath := filepath.Join(path, ".tmsu", "db")

		log.Infof(2, "looking for database at '%s'", dbPath)

		_, err := os.Stat(dbPath)
		if err == nil {
			return dbPath, nil
		}

		switch {
		case os.IsNotExist(err):
			if _path.IsRoot(path) {
				return "", nil
			}

			path = filepath.Dir(path)
			continue
		case os.IsPermission(err):
			return "", nil
		default:
			return "", err
		}
	}
}

func findCommand(commands []*Command, commandName string) *Command {
	for _, command := range commands {
		if command.Name == commandName {
			return command
		}

		for _, alias := range command.Aliases {
			if alias == commandName {
				return command
			}
		}

	}

	return nil
}
