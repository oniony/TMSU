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
	"fmt"
	"sort"
	"tmsu/common"
	"tmsu/database"
)

type FilesCommand struct{}

func (FilesCommand) Name() string {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "List files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files [--explicit] [-]TAG...
tmsu files --all

Lists the files, if any, that have all of the TAGs specified. Tags can be excluded by prefixing them with a minus (-).

  --all         show the complete set of tagged files
  --explicit    show only files tagged explicitly`
}

func (command FilesCommand) Exec(args []string) error {
	argCount := len(args)

	if argCount == 1 && args[0] == "--all" {
		return command.listAllFiles()
	}

	explicit := false
	if argCount > 0 && args[0] == "--explicit" {
		explicit = true
		args = args[1:]
	}

	return command.listFiles(args, explicit)
}

func (FilesCommand) listAllFiles() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	files, err := db.Files()
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Println(file.Path())
	}

	return nil
}

func (FilesCommand) listFiles(args []string, explicit bool) error {
	if len(args) == 0 {
		return errors.New("At least one tag must be specified. Use --all to show all files.")
	}

	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	includeTagIds := make([]uint, 0)
	excludeTagIds := make([]uint, 0)
	for _, arg := range args {
		var tagName string
		var include bool
		if arg[0] == '-' {
			tagName = arg[1:]
			include = false
		} else {
			tagName = arg
			include = true
		}

		tag, err := db.TagByName(tagName)
		if err != nil {
			return err
		}
		if tag == nil {
			return errors.New("No such tag '" + tagName + "'.")
		}

		if include {
			includeTagIds = append(includeTagIds, tag.Id)
		} else {
			excludeTagIds = append(excludeTagIds, tag.Id)
		}
	}

	files, err := db.FilesWithTags(includeTagIds, excludeTagIds, explicit)
	if err != nil {
		return err
	}

	paths := make([]string, len(files))
	for _, file := range files {
		relPath := common.MakeRelative(file.Path())
		paths = append(paths, relPath)
	}

	sort.Strings(paths)

	previousPath := ""
	for _, path := range paths {
		if path != previousPath {
			fmt.Println(path)
		}

		previousPath = path
	}

	return nil
}
