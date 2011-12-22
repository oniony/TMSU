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

func (this UntagCommand) Name() string {
	return "untag"
}

func (this UntagCommand) Summary() string {
	return "removes all tags or specific tags from a file"
}

func (this UntagCommand) Help() string {
	return `  tmsu untag FILE TAG...
  tmsu untag --all FILE...

Disassociates FILE with the TAGs specified.

  --all    strip each FILE of all tags`
}

func (this UntagCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("No arguments specified.")
	}

    if args[0] == "--all" {
        if len(args) < 2 { return errors.New("Files to untag must be specified.") }

        error := this.removeFiles(args[1:])
        if error != nil { return error }
    } else {
        if len(args) < 2 { return errors.New("Tags to remove must be specified.") }

        error := this.untagFile(args[0], args[1:])
        if error != nil { return error }
    }


	return nil
}

// implementation

func (this UntagCommand) removeFiles(paths []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil {
        return error
    }
    defer db.Close()

    for _, path := range paths {
        absPath, error := filepath.Abs(path)
        if error != nil {
            return error
        }

        file, error := db.FileByPath(absPath)
        if error != nil { return error }
        if file == nil { return errors.New("File '" + path + "' is not tagged.") }

        error = db.RemoveFileTagsByFileId(file.Id)
        if error != nil { return error }

        error = db.RemoveFile(file.Id)
        if error != nil { return error }
    }

    return nil
}

func (this UntagCommand) untagFile(path string, tagNames []string) error {
	absPath, error := filepath.Abs(path)
	if error != nil { return error }

	db, error := OpenDatabase(databasePath())
	if error != nil { return error }
	defer db.Close()

	file, error := db.FileByPath(absPath)
	if error != nil { return error }
	if file == nil { return errors.New("File '" + path + "' is not tagged.") }

    for _, tagName := range tagNames {
        error = this.unapplyTag(db, path, file.Id, tagName)
        if error != nil { return error }
    }

	hasTags, error := db.AnyFileTagsForFile(file.Id)
	if error != nil { return error }

	if !hasTags {
        error := db.RemoveFile(file.Id)
        if error != nil { return error }
	}

	return nil
}

func (this UntagCommand) unapplyTag(db *Database, path string, fileId uint, tagName string) error {
	tag, error := db.TagByName(tagName)
	if error != nil {
		return error
	}
	if tag == nil {
		errors.New("No such tag" + tagName)
	}

	fileTag, error := db.FileTagByFileIdAndTagId(fileId, tag.Id)
	if error != nil {
		return error
	}
	if fileTag == nil {
		errors.New("File '" + path + "' is not tagged '" + tagName + "'.")
	}

	if fileTag != nil {
		error := db.RemoveFileTag(fileId, tag.Id)
		if error != nil {
			return error
		}
	} else {
		return errors.New("File '" + path + "' is not tagged '" + tagName + "'.\n")
	}

	return nil
}
