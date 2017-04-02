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
	"github.com/oniony/TMSU/storage"
)

var DeleteCommand = Command{
	Name:        "delete",
	Aliases:     []string{"del", "rm"},
	Synopsis:    "Delete one or more tags",
	Usages:      []string{"tmsu delete TAG..."},
	Description: `Permanently deletes the TAGs specified.`,
	Examples: []string{"$ tmsu delete pineapple",
		"$ tmsu delete red green blue"},
	Options: Options{Option{"--value", "", "delete a value", false, ""}},
	Exec:    deleteExec,
}

// unexported

func deleteExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) == 0 {
		return fmt.Errorf("too few arguments"), nil
	}

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
		return deleteValue(store, tx, args)
	}

	return deleteTag(store, tx, args)
}

func deleteTag(store *storage.Storage, tx *storage.Tx, tagArgs []string) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, tagArg := range tagArgs {
		tagName := parseTagOrValueName(tagArg)

		tag, err := store.TagByName(tx, tagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err), warnings
		}
		if tag == nil {
			warnings = append(warnings, fmt.Sprintf("no such tag '%v'", tagName))
			continue
		}

		err = store.DeleteTag(tx, tag.Id)
		if err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", tagName, err), warnings
		}
	}

	return nil, warnings
}

func deleteValue(store *storage.Storage, tx *storage.Tx, valueArgs []string) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, valueArg := range valueArgs {
		valueName := parseTagOrValueName(valueArg)

		value, err := store.ValueByName(tx, valueName)
		if err != nil {
			return fmt.Errorf("could not retrieve value '%v': %v", valueName, err), warnings
		}
		if value == nil {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", valueName))
			continue
		}

		if err = store.DeleteValue(tx, value.Id); err != nil {
			return fmt.Errorf("could not delete value '%v': %v", valueName, err), warnings
		}
	}

	return nil, warnings
}
