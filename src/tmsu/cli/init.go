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
	"path/filepath"
	"tmsu/common/log"
	"tmsu/storage"
)

var InitCommand = Command{
	Name:     "init",
	Synopsis: "Initializes a new database",
	Usages:   []string{"tmsu init [PATH]"},
	Description: `Initializes a new local database.

Creates a .tmsu directory under PATH and initialises a new empty database within it.

If PATH is omitted then the current working directory is assumed.

The new database is used automatically whenever TMSU is invoked from a directory under PATH (unless overriden by the global --database option or the TMSU_DB environment variable.`,
	Options: Options{},
	Exec:    initExec,
}

func initExec(store *storage.Storage, options Options, args []string) error {
	paths := args

	if len(paths) == 0 {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not identify working directory: %v", err)
		}

		paths = []string{workingDirectory}
	}

    if err := store.Begin(); err != nil {
        return err
    }
    defer store.Commit()

	wereErrors := false
	for _, path := range paths {
		if err := initializeDatabase(path); err != nil {
			log.Warnf("%v: could not initialize database", path, err)
			wereErrors = true
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

// unexported

func initializeDatabase(path string) error {
	tmsuPath := filepath.Join(path, ".tmsu")

	if err := os.Mkdir(tmsuPath, 0755); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("could not create .tmsu directory: %v", err)
		}
	}

	dbPath := filepath.Join(tmsuPath, "db")

	store, err := storage.OpenAt(dbPath)
	if err != nil {
		return fmt.Errorf("%v: could not open database", dbPath, err)
	}
	store.Close()

	return nil
}
