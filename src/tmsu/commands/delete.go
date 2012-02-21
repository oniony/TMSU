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

type DeleteCommand struct{}

func (DeleteCommand) Name() string {
	return "delete"
}

func (DeleteCommand) Synopsis() string {
	return "Delete one or more tags"
}

func (DeleteCommand) Description() string {
	return `tmsu delete TAG...

Permanently deletes the TAGs specified.`
}

func (command DeleteCommand) Exec(args []string) error {
	if len(args) == 0 {
		return errors.New("No tags to delete specified.")
	}

	db, err := database.OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, tagName := range args {
		err = command.deleteTag(db, tagName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (DeleteCommand) deleteTag(db *database.Database, tagName string) error {
	tag, err := db.TagByName(tagName)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	fileTags, err := db.FileTagsByTagId(tag.Id)
	if err != nil {
		return err
	}

	err = db.RemoveFileTagsByTagId(tag.Id)
	if err != nil {
		return err
	}

	err = db.DeleteTag(tag.Id)
	if err != nil {
		return err
	}

	for _, fileTag := range fileTags {
		hasTags, err := db.AnyFileTagsForFile(fileTag.FileId)
		if err != nil {
			return err
		}

		if !hasTags {
			db.RemoveFile(fileTag.FileId)
		}
	}

	return nil
}
