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
)

var RenameCommand = Command{
	Name:     "rename",
	Aliases:  []string{"mv"},
	Synopsis: "Rename a tag or value",
	Usages:   []string{"tmsu rename [OPTION]... OLD NEW"},
	Description: `Renames a tag or value from OLD to NEW.

Attempting to rename a tag or value with a name that already exists will result in an error. To merge tags or values use the 'merge' subcommand instead.`,
	Examples: []string{"$ tmsu rename montain mountain",
		"$ tmsu rename --value MMXVII 2017"},
	Options: Options{{"--value", "", "rename a value", false, ""}},
	Exec:    renameExec,
}

// unexported

func renameExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments"), nil
	}

	if len(args) > 2 {
		return fmt.Errorf("too many arguments"), nil
	}

	currentName := parseTagOrValueName(args[0])
	newName := parseTagOrValueName(args[1])

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	if options.HasOption("--value") {
		return renameValue(store, tx, currentName, newName), nil
	}

	return renameTag(store, tx, currentName, newName), nil
}

func renameTag(store *storage.Storage, tx *storage.Tx, currentName, newName string) error {
	sourceTag, err := store.TagByName(tx, currentName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", currentName, err)
	}
	if sourceTag == nil {
		return fmt.Errorf("no such tag '%v'", currentName)
	}

	destTag, err := store.TagByName(tx, newName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", newName, err)
	}
	if destTag != nil {
		return fmt.Errorf("tag '%v' already exists", newName)
	}

	log.Infof(2, "renaming tag '%v' to '%v'.", currentName, newName)

	_, err = store.RenameTag(tx, sourceTag.Id, newName)
	if err != nil {
		return fmt.Errorf("could not rename tag '%v' to '%v': %v", currentName, newName, err)
	}

	return nil
}

func renameValue(store *storage.Storage, tx *storage.Tx, currentName, newName string) error {
	sourceValue, err := store.ValueByName(tx, currentName)
	if err != nil {
		return fmt.Errorf("could not retrieve value '%v': %v", currentName, err)
	}
	if sourceValue == nil {
		return fmt.Errorf("no such value '%v'", currentName)
	}

	destValue, err := store.ValueByName(tx, newName)
	if err != nil {
		return fmt.Errorf("could not retrieve value '%v': %v", newName, err)
	}
	if destValue != nil {
		return fmt.Errorf("value '%v' already exists", newName)
	}

	log.Infof(2, "renaming value '%v' to '%v'.", currentName, newName)

	_, err = store.RenameValue(tx, sourceValue.Id, newName)
	if err != nil {
		return fmt.Errorf("could not rename value '%v' to '%v': %v", currentName, newName, err)
	}

	return nil
}
