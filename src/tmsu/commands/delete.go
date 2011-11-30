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
package main
*/

package main

import (
	"errors"
)

type DeleteCommand struct{}

func (this DeleteCommand) Name() string {
	return "delete"
}

func (this DeleteCommand) Summary() string {
	return "deletes one or more tags"
}

func (this DeleteCommand) Help() string {
	return `tmsu delete TAG...

Permanently deletes the tag(s) specified.`
}

func (this DeleteCommand) Exec(args []string) error {
	if len(args) == 0 {
		return errors.New("No tags to delete specified.")
	}

	db, error := OpenDatabase(databasePath())
	if error != nil {
		return error
	}
	defer db.Close()

	for _, tagName := range args {
		error = this.deleteTag(db, tagName)
		if error != nil {
			return error
		}
	}

	return nil
}

func (this DeleteCommand) deleteTag(db *Database, tagName string) error {
	tag, error := db.TagByName(tagName)
	if error != nil {
		return error
	}

	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	fileTags, error := db.FileTagsByTagId(tag.Id)
	if error != nil {
		return error
	}

	error = db.RemoveFileTagsByTagId(tag.Id)
	if error != nil {
		return error
	}

	error = db.DeleteTag(tag.Id)
	if error != nil {
		return error
	}

	for _, fileTag := range *fileTags {
		hasTags, error := db.AnyFileTagsForFile(fileTag.FileId)
		if error != nil {
			return error
		}

		if !hasTags {
			db.RemoveFile(fileTag.FileId)
		}
	}

	return nil
}
