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
	"fmt"
	"sort"
)

type FilesCommand struct{}

func (FilesCommand) Name() string {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "lists files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files TAG...
tmsu files --all

Lists the files, if any, that have all of the TAGs specified.

  --all    show the complete set of tagged files`
}

func (command FilesCommand) Exec(args []string) error {
	argCount := len(args)

	if argCount == 1 && args[0] == "--all" {
		return command.listAllFiles()
	}

	return command.listFiles(args)
}

func (FilesCommand) listAllFiles() error {
	db, err := OpenDatabase()
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

func (FilesCommand) listFiles(tagNames []string) error {
	if len(tagNames) == 0 {
		return errors.New("At least one tag must be specified.")
	}

	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, tagName := range tagNames {
		tag, err := db.TagByName(tagName)
		if err != nil {
			return err
		}
		if tag == nil {
			return errors.New("No such tag '" + tagName + "'.")
		}
	}

	files, err := db.FilesWithTags(tagNames)
	if err != nil {
		return err
	}

	paths := make([]string, len(files))
	for index, file := range files {
		relPath := makeRelative(file.Path())
		paths[index] = relPath
	}

	sort.Strings(paths)
	for _, path := range paths {
		fmt.Println(path)
	}

	return nil
}
