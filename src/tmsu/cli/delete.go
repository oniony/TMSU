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
	"tmsu/api"
	"tmsu/common/log"
	"tmsu/storage"
)

var DeleteCommand = Command{
	Name:        "delete",
	Aliases:     []string{"del", "rm"},
	Synopsis:    "Delete one or more tags",
	Usages:      []string{"tmsu delete TAG..."},
	Description: `Permanently deletes the TAGs specified.`,
	Examples: []string{"$ tmsu delete pineapple",
		"$ tmsu delete red green blue"},
	Options: Options{},
	Exec:    deleteExec,
}

func deleteExec(store *storage.Storage, options Options, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("too few arguments")
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	wereErrors := false
	for _, tagName := range args {
		err = api.DeleteTag(store, tx, tagName)
		if err != nil {
			switch err.(type) {
			case api.NoSuchTag:
				log.Warn(err.Error())
				wereErrors = true
			default:
				return fmt.Errorf("could not delete tag '%v': %v", tagName, err)
			}
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}
