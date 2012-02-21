/*
Copyright 2011-2012 Paul Ruane.

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

package commands

import (
	"errors"
	"tmsu/database"
)

type CopyCommand struct{}

func (CopyCommand) Name() string {
	return "copy"
}

func (CopyCommand) Synopsis() string {
	return "Create a copy of a tag"
}

func (CopyCommand) Description() string {
	return `tmsu copy TAG NEW

Creates a new tag NEW applied to the same set of files as TAG.`
}

func (CopyCommand) Exec(args []string) error {
	db, err := database.OpenDatabase()
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

	_, err = db.CopyTag(sourceTag.Id, destTagName)
	if err != nil {
		return err
	}

	return nil
}
