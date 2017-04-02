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
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
)

var InitCommand = Command{
	Name:     "init",
	Synopsis: "Initializes a new database",
	Usages:   []string{"tmsu init [PATH]"},
	Description: `Initializes a new local database.

Creates a .tmsu directory under PATH and initialises a new empty database within it.

If no PATH is specified then the current working directory is assumed.

The new database is used automatically whenever TMSU is invoked from a directory under PATH (unless overridden by the global --database option or the TMSU_DB environment variable.`,
	Options: Options{},
	Exec:    initExec,
}

// unexported

func initExec(options Options, args []string, databasePath string) (error, warnings) {
	paths := args

	if len(paths) == 0 {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not identify working directory: %v", err), nil
		}

		paths = []string{workingDirectory}
	}

	warnings := make(warnings, 0, 10)
	for _, path := range paths {
		if err := initializeDatabase(path); err != nil {
			warnings = append(warnings, fmt.Sprintf("%v: could not initialize database: %v", path, err))
		}
	}

	return nil, warnings
}

func initializeDatabase(path string) error {
	log.Warnf("%v: creating database", path)

	tmsuPath := filepath.Join(path, ".tmsu")
	os.Mkdir(tmsuPath, 0755)

	dbPath := filepath.Join(tmsuPath, "db")

	return storage.CreateAt(dbPath)
}
