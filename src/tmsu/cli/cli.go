// Copyright 2011-2015 Paul Ruane.

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
	"os"
	"os/user"
	"path/filepath"
	"tmsu/common/log"
	_path "tmsu/common/path"
	"tmsu/storage"
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
		// find version command
	case options.HasOption("--help"), command == nil:
		// find help command
	}

	log.Verbosity = options.Count("--verbose") + 1

	var databasePath string
	switch {
	case options.HasOption("--database"):
		databasePath = options.Get("--database").Argument
	case os.Getenv("TMSU_DB") != "":
		databasePath = os.Getenv("TMSU_DB")
	default:
		databasePath, err = findDatabase()
		if err != nil {
			log.Fatalf("could not find database: %v", err)
		}
	}

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		log.Fatalf("could not open storage: %v", err)
	}

	if err = processCommand(store, command, options, arguments); err != nil {
		if err != errBlank {
			log.Warn(err.Error())
		}

		store.Close()
		os.Exit(1)
	}

	store.Close()
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

func processCommand(store *storage.Storage, command *Command, options Options, arguments []string) error {
	if err := command.Exec(store, options, arguments); err != nil {
		return err
	}

	return nil
}
