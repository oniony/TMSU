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
	"path/filepath"
)

type UntagCommand struct{}

func (UntagCommand) Name() string {
	return "untag"
}

func (UntagCommand) Summary() string {
	return "removes all tags or specific tags from a file"
}

func (UntagCommand) Help() string {
	return `  tmsu untag FILE TAG...
  tmsu untag --all FILE...

Disassociates FILE with the TAGs specified.

  --all    strip each FILE of all tags`
}

func (command UntagCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("No arguments specified.")
	}

    if args[0] == "--all" {
        if len(args) < 2 { return errors.New("Files to untag must be specified.") }

        err := command.removeFiles(args[1:])
        if err != nil { return err }
    } else {
        if len(args) < 2 { return errors.New("Tags to remove must be specified.") }

        err := command.untagFile(args[0], args[1:])
        if err != nil { return err }
    }


	return nil
}

func (UntagCommand) removeFiles(paths []string) error {
    db, err := OpenDatabase(databasePath())
    if err != nil { return err }
    defer db.Close()

    for _, path := range paths {
        absPath, err := filepath.Abs(path)
        if err != nil { return err }

        file, err := db.FileByPath(absPath)
        if err != nil { return err }
        if file == nil { return errors.New("File '" + path + "' is not tagged.") }

        err = db.RemoveFileTagsByFileId(file.Id)
        if err != nil { return err }

        err = db.RemoveFile(file.Id)
        if err != nil { return err }
    }

    return nil
}

func (command UntagCommand) untagFile(path string, tagNames []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil { return err }

	db, err := OpenDatabase(databasePath())
	if err != nil { return err }
	defer db.Close()

	file, err := db.FileByPath(absPath)
	if err != nil { return err }
	if file == nil { return errors.New("File '" + path + "' is not tagged.") }

    for _, tagName := range tagNames {
        err = command.unapplyTag(db, path, file.Id, tagName)
        if err != nil { return err }
    }

	hasTags, err := db.AnyFileTagsForFile(file.Id)
	if err != nil { return err }

	if !hasTags {
        err := db.RemoveFile(file.Id)
        if err != nil { return err }
	}

	return nil
}

func (UntagCommand) unapplyTag(db *Database, path string, fileId uint, tagName string) error {
	tag, err := db.TagByName(tagName)
	if err != nil { return err }
	if tag == nil {
		errors.New("No such tag" + tagName)
	}

	fileTag, err := db.FileTagByFileIdAndTagId(fileId, tag.Id)
	if err != nil { return err }
	if fileTag == nil { return errors.New("File '" + path + "' is not tagged '" + tagName + "'.") }

	if fileTag != nil {
		err := db.RemoveFileTag(fileId, tag.Id)
		if err != nil { return err }
	} else {
		return errors.New("File '" + path + "' is not tagged '" + tagName + "'.\n")
	}

	return nil
}
