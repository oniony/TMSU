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

func (RenameCommand) Name() string {
	return "rename"
}

func (RenameCommand) Summary() string {
	return "renames a tag"
}

func (RenameCommand) Help() string {
	return `tmsu rename OLD NEW

Renames a tag from OLD to NEW.

Attempting to rename a tag with a new name for which a tag already exists will result in an error.
To merge tags use the 'merge' command instead.`
}

func (RenameCommand) Exec(args []string) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, err := db.TagByName(sourceTagName)
	if err != nil {
		return err
	}
	if sourceTag == nil {
		return errors.New("No such tag '" + sourceTagName + "'.")
	}

	destTag, err := db.TagByName(destTagName)
	if err != nil {
		return err
	}
	if destTag != nil {
		return errors.New("A tag with name '" + destTagName + "' already exists.")
	}

	_, err = db.RenameTag(sourceTag.Id, destTagName)
	if err != nil {
		return err
	}

	return nil
}
