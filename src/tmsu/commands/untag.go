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
	"path/filepath"
	"strings"
	"tmsu/common"
	"tmsu/database"
)

type UntagCommand struct{}

func (UntagCommand) Name() string {
	return "untag"
}

func (UntagCommand) Synopsis() string {
	return "Remove tags from files"
}

func (UntagCommand) Description() string {
	return `tmsu untag FILE TAG...
tmsu untag --all FILE...
tmsu untag --tags "TAG..." FILE...

Disassociates FILE with the TAGs specified.

  --all     strip each FILE of all TAGs
  --tags    disassociate multiple FILEs from the same quoted set of TAGs`
}

func (command UntagCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("No arguments specified.")
	}

	switch args[0] {
	case "--all":
		if len(args) < 2 {
			return errors.New("Files to untag must be specified.")
		}

		err := command.removeFiles(args[1:])
		if err != nil {
			return err
		}
	case "--tags":
		if len(args) < 3 {
			return errors.New("Quoted set of tags and at least one file to untag must be specified.")
		}

		tagNames := strings.Fields(args[1])
		paths := args[2:]

		err := command.untagPaths(paths, tagNames)
		if err != nil {
			return err
		}
	default:
		if len(args) < 2 {
			return errors.New("Tags to remove must be specified.")
		}

		err := command.untagPath(args[0], args[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (UntagCommand) removeFiles(paths []string) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		file, err := db.FileByPath(absPath)
		if err != nil {
			return err
		}
		if file == nil {
			common.Warnf("'%v': file is not tagged", path)
			continue
		}

		err = db.RemoveFileTagsByFileId(file.Id)
		if err != nil {
			return err
		}

		err = db.RemoveFile(file.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPaths(paths []string, tagNames []string) error {
	for _, path := range paths {
		err := command.untagPath(path, tagNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPath(path string, tagNames []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	file, err := db.FileByPath(absPath)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("File '" + path + "' is not tagged.")
	}

	for _, tagName := range tagNames {
		err = command.unapplyTag(db, path, file.Id, tagName)
		if err != nil {
			return err
		}
	}

	hasTags, err := db.AnyFileTagsForFile(file.Id)
	if err != nil {
		return err
	}

	if !hasTags {
		err := db.RemoveFile(file.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (UntagCommand) unapplyTag(db *database.Database, path string, fileId uint, tagName string) error {
	tag, err := db.TagByName(tagName)
	if err != nil {
		return err
	}
	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	fileTag, err := db.FileTagByFileIdAndTagId(fileId, tag.Id)
	if err != nil {
		return err
	}
	if fileTag == nil {
		return errors.New("File '" + path + "' is not tagged '" + tagName + "'.")
	}

    err = db.RemoveFileTag(fileId, tag.Id)
    if err != nil {
        return err
    }

	return nil
}
