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
	"tmsu/cli"
	"tmsu/storage"
)

type DeleteCommand struct{}

func (DeleteCommand) Name() cli.CommandName {
	return "delete"
}

func (DeleteCommand) Synopsis() string {
	return "Delete one or more tags"
}

func (DeleteCommand) Description() string {
	return `tmsu delete TAG...

Permanently deletes the TAGs specified.`
}

func (DeleteCommand) Options() cli.Options {
	return cli.Options{}
}

func (command DeleteCommand) Exec(options cli.Options, args []string) error {
	if len(args) == 0 {
		return errors.New("No tags to delete specified.")
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	for _, tagName := range args {
		err = command.deleteTag(store, tagName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (DeleteCommand) deleteTag(store *storage.Storage, tagName string) error {
	tag, err := store.TagByName(tagName)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	fileTags, err := store.FileTagsByTagId(tag.Id, true)
	if err != nil {
		return err
	}

	err = store.RemoveFileTagsByTagId(tag.Id)
	if err != nil {
		return err
	}

	err = store.DeleteTag(tag.Id)
	if err != nil {
		return err
	}

	for _, fileTag := range fileTags {
		tags, err := store.TagsByFileId(fileTag.FileId, false)
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			store.RemoveFile(fileTag.FileId)
		}
	}

	return nil
}
