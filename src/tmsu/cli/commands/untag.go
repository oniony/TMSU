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
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type UntagCommand struct{}

func (UntagCommand) Name() cli.CommandName {
	return "untag"
}

func (UntagCommand) Synopsis() string {
	return "Remove tags from files"
}

func (UntagCommand) Description() string {
	return `tmsu untag FILE TAG...
tmsu untag --all FILE...
tmsu untag --tags "TAG..." FILE...

Disassociates FILE with the TAGs specified.`
}

func (UntagCommand) Options() cli.Options {
	return cli.Options{{"-a", "--all", "strip each file of all tags"},
		{"-t", "--tags", "the set of tags to remove"}}
}

func (command UntagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return errors.New("No arguments specified.")
	}

	if cli.HasOption(options, "--all") {
		if len(args) < 1 {
			return errors.New("Files to untag must be specified.")
		}

		err := command.untagAllPaths(args)
		if err != nil {
			return err
		}
	} else if cli.HasOption(options, "--tags") {
		if len(args) < 2 {
			return errors.New("Quoted set of tags and at least one file to untag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		err := command.untagPaths(paths, tagNames)
		if err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return errors.New("Tag to remove and files to untag must be specified.")
		}

		err := command.untagPath(args[0], args[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func (UntagCommand) untagAllPaths(paths []string) error {
	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return err
		}
		if file == nil {
			log.Warnf("'%v': file is not tagged.", path)
			continue
		}

		removeFile := true
		fileTags, err := store.FileTagsByFileId(file.Id, false)
		for _, fileTag := range fileTags {
			if fileTag.Explicit {
				if !fileTag.Implicit {
					err = store.RemoveFileTag(fileTag.Id)
					if err != nil {
						return err
					}

					//TODO recursively remove implicit file tags beneath
				} else {
					err = store.UpdateFileTag(fileTag.Id, false, true)
					if err != nil {
						return err
					}
				}
			} else {
				removeFile = false
			}
		}

		if removeFile {
			err = store.RemoveFile(file.Id)
			if err != nil {
				return err
			}
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

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	file, err := store.FileByPath(absPath)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("File '" + path + "' is not tagged.")
	}

	for _, tagName := range tagNames {
		err = command.unapplyTag(store, path, file.Id, tagName)
		if err != nil {
			return err
		}
	}

	filetags, err := store.FileTagsByFileId(file.Id, false)
	if err != nil {
		return err
	}

	if len(filetags) == 0 {
		err := store.RemoveFile(file.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (UntagCommand) unapplyTag(store *storage.Storage, path string, fileId uint, tagName string) error {
	tag, err := store.TagByName(tagName)
	if err != nil {
		return err
	}
	if tag == nil {
		return errors.New("No such tag '" + tagName + "'.")
	}

	fileTag, err := store.FileTagByFileIdAndTagId(fileId, tag.Id)
	if err != nil {
		return err
	}
	if fileTag == nil {
		return errors.New("File '" + path + "' is not tagged '" + tagName + "'.")
	}

	err = store.RemoveFileTag(fileTag.Id)
	if err != nil {
		return err
	}

	return nil
}
