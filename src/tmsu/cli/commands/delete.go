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

package commands

import (
	"errors"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type DeleteCommand struct {
	verbose bool
}

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

	command.verbose = options.HasOption("--verbose")

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

func (command DeleteCommand) deleteTag(store *storage.Storage, tagName string) error {
	tag, err := store.TagByName(tagName)
	if err != nil {
		return err
	}
	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	if command.verbose {
		log.Infof("finding files explicitly tagged '%v'.", tagName)
	}

	fileTags, err := store.FileTagsByTagId(tag.Id)
	if err != nil {
		return err
	}

	if command.verbose {
		log.Infof("removing applications of tag '%v'.", tagName)
	}

	err = store.RemoveFileTagsByTagId(tag.Id)
	if err != nil {
		return err
	}

	if command.verbose {
		log.Infof("deleting tag '%v'.", tagName)
	}

	err = store.DeleteTag(tag.Id)
	if err != nil {
		return err
	}

	if command.verbose {
		log.Infof("identifying files left untagged as a result of tag deletion.")
	}

	removedFileCount := 0
	for _, fileTag := range fileTags {
		count, err := store.FileTagCountByFileId(fileTag.FileId)
		if err != nil {
			return err
		}
		if count == 0 {
			err := store.RemoveFile(fileTag.FileId)
			if err != nil {
				return err
			}

			removedFileCount += 1
		}
	}

	if command.verbose {
		log.Infof("removed %v untagged files.", removedFileCount)
	}

	return nil
}
