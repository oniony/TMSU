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
	"tmsu/common/log"
	"tmsu/storage"
)

var RenameCommand = Command{
	Name:     "rename",
	Aliases:  []string{"mv"},
	Synopsis: "Rename a tag",
	Usages:   []string{"tmsu rename OLD NEW"},
	Description: `Renames a tag from OLD to NEW.

Attempting to rename a tag with a new name for which a tag already exists will result in an error. To merge tags use the 'merge' subcommand instead.`,
	Examples: []string{"$ tmsu rename montain mountain"},
	Options:  Options{},
	Exec:     renameExec,
}

func renameExec(store *storage.Storage, options Options, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments")
	}

	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}

	sourceTagName := args[0]
	destTagName := args[1]

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	sourceTag, err := store.TagByName(tx, sourceTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
	}
	if sourceTag == nil {
		return fmt.Errorf("no such tag '%v'", sourceTagName)
	}

	destTag, err := store.TagByName(tx, destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
	}
	if destTag != nil {
		return fmt.Errorf("tag '%v' already exists", destTagName)
	}

	log.Infof(2, "renaming tag '%v' to '%v'.", sourceTagName, destTagName)

	_, err = store.RenameTag(tx, sourceTag.Id, destTagName)
	if err != nil {
		return fmt.Errorf("could not rename tag '%v' to '%v': %v", sourceTagName, destTagName, err)
	}

	return nil
}
