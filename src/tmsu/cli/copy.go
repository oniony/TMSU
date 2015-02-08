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

var CopyCommand = Command{
	Name:        "copy",
	Aliases:     []string{"cp"},
	Synopsis:    "Create a copy of a tag",
	Usages:      []string{"tmsu copy TAG NEW..."},
	Description: `Creates a new tag NEW applied to the same set of files as TAG.`,
	Examples: []string{"$ tmsu copy cheese wine",
		"$ tmsu copy report document text"},
	Options: Options{},
	Exec:    copyExec,
}

func copyExec(store *storage.Storage, options Options, args []string) error {
    if len(args) < 2 {
        return fmt.Errorf("the tag to copy and at least one copy name must be specified")
    }

	sourceTagName := args[0]
	destTagNames := args[1:]

	if err := store.Begin(); err != nil {
		return err
	}
	defer store.Commit()

	sourceTag, err := store.Db.TagByName(sourceTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
	}
	if sourceTag == nil {
		return fmt.Errorf("no such tag '%v'", sourceTagName)
	}

	wereErrors := false
	for _, destTagName := range destTagNames {
		destTag, err := store.Db.TagByName(destTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
		}
		if destTag != nil {
			log.Warnf("a tag with name '%v' already exists.", destTagName)
			wereErrors = true
			continue
		}

		log.Infof(2, "copying tag '%v' to '%v'.", sourceTagName, destTagName)

		if _, err = store.CopyTag(sourceTag.Id, destTagName); err != nil {
			return fmt.Errorf("could not copy tag '%v' to '%v': %v", sourceTagName, destTagName, err)
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}
