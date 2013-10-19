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
	"tmsu/log"
	"tmsu/storage"
)

var DeleteCommand = &Command{
	Name:     "delete",
	Synopsis: "Delete one or more tags",
	Description: `tmsu delete TAG...

Permanently deletes the TAGs specified.`,
	Options: Options{},
	Exec:    deleteExec,
}

func deleteExec(options Options, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no tags to delete specified.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	for _, tagName := range args {
		err = deleteTag(store, tagName)
		if err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", tagName, err)
		}
	}

	return nil
}

func deleteTag(store *storage.Storage, tagName string) error {
	tag, err := store.TagByName(tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return fmt.Errorf("no such tag '%v'.", tagName)
	}

	log.Suppf("finding files tagged '%v'.", tagName)

	fileTags, err := store.FileTagsByTagId(tag.Id)
	if err != nil {
		return fmt.Errorf("could not retrieve taggings for tag '%v': %v", tagName, err)
	}

	log.Suppf("removing applications of tag '%v'.", tagName)

	err = store.RemoveFileTagsByTagId(tag.Id)
	if err != nil {
		return fmt.Errorf("could not remove taggings for tag '%v': %v", tagName, err)
	}

	log.Suppf("removing tags implications involving tag '%v'.", tagName)

	err = store.RemoveImplicationsForTagId(tag.Id)
	if err != nil {
		return fmt.Errorf("could not remove tag implications involving tag '%v': %v", tagName, err)
	}

	log.Suppf("deleting tag '%v'.", tagName)

	err = store.DeleteTag(tag.Id)
	if err != nil {
		return fmt.Errorf("could not delete tag '%v': %v", tagName, err)
	}

	log.Suppf("identifying files left untagged as a result of tag deletion.")

	removedFileCount := 0
	for _, fileTag := range fileTags {
		count, err := store.FileTagCountByFileId(fileTag.FileId)
		if err != nil {
			return fmt.Errorf("could not retrieve taggings count for file #%v: %v", fileTag.FileId, err)
		}
		if count == 0 {
			err := store.RemoveFile(fileTag.FileId)
			if err != nil {
				return fmt.Errorf("could not remove file #%v: %v", fileTag.FileId, err)
			}

			removedFileCount += 1
		}
	}

	log.Suppf("removed %v untagged files.", removedFileCount)

	return nil
}
