/*
Copyright 2011 Paul Ruane.

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

package main

import (
	"errors"
)

type RenameCommand struct{}

func (this RenameCommand) Name() string {
	return "rename"
}

func (this RenameCommand) Summary() string {
	return "renames a tag"
}

func (this RenameCommand) Help() string {
	return `  tmsu rename OLD NEW

Renames a tag from OLD to NEW.

Attempting to rename a tag with a new name for which a tag already exists will result in an error.
To merge tags use the 'merge' command instead.`
}

func (this RenameCommand) Exec(args []string) error {
	db, error := OpenDatabase(databasePath())
	if error != nil { return error }
	defer db.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, error := db.TagByName(sourceTagName)
	if error != nil { return error }
	if sourceTag == nil {
		return errors.New("No such tag '" + sourceTagName + "'.")
	}

	destTag, error := db.TagByName(destTagName)
	if error != nil { return error }
	if destTag != nil {
		return errors.New("A tag with name '" + destTagName + "' already exists.")
	}

	_, error = db.RenameTag(sourceTag.Id, destTagName)
	if error != nil { return error }

	return nil
}
