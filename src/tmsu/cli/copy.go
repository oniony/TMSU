/*
Copyright 2011-2013 Paul Ruane.

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
	"fmt"
	"tmsu/common/log"
	"tmsu/storage"
)

var CopyCommand = Command{
	Name:     "copy",
	Synopsis: "Create a copy of a tag",
	Description: `tmsu copy TAG NEW...

Creates a new tag NEW applied to the same set of files as TAG.

Examples:

    $ tmsu copy cheese wine
    $ tmsu copy report document`,
	Options: Options{},
	Exec:    copyExec,
}

func copyExec(options Options, args []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	sourceTagName := args[0]
	destTagNames := args[1:]

	sourceTag, err := store.Db.TagByName(sourceTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
	}
	if sourceTag == nil {
		return fmt.Errorf("no such tag '%v'.", sourceTagName)
	}

	for _, destTagName := range destTagNames {
		destTag, err := store.Db.TagByName(destTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
		}
		if destTag != nil {
			return fmt.Errorf("a tag with name '%v' already exists.", destTagName)
		}

		log.Infof(2, "copying tag '%v' to '%v'.", sourceTagName, destTagName)

		if _, err = store.CopyTag(sourceTag.Id, destTagName); err != nil {
			return fmt.Errorf("could not copy tag '%v' to '%v': %v", sourceTagName, destTagName, err)
		}
	}

	return nil
}
