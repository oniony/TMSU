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
	"fmt"
	"tmsu/cli"
	"tmsu/common"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type FilesCommand struct {
	verbose bool
}

func (FilesCommand) Name() cli.CommandName {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "List files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files [--explicit] [-]TAG...
tmsu files --all

Lists the files, if any, that have all of the TAGs specified. Tags can be excluded by prefixing them with a minus (-).`
}

func (FilesCommand) Options() cli.Options {
	return cli.Options{{"-a", "--all", "show the complete set of tagged files"},
		{"-e", "--explicit", "show only the explicitly tagged files"}}
}

func (command FilesCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	if options.HasOption("--all") {
		return command.listAllFiles()
	}

	explicitOnly := options.HasOption("--explicit")

	return command.listFiles(args, explicitOnly)
}

func (command FilesCommand) listAllFiles() error {
	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	if command.verbose {
		log.Info("retrieving all files from database.")
	}

	files, err := store.Files()
	if err != nil {
		return err
	}

	for _, file := range files {
		relPath := common.RelPath(file.Path())
		fmt.Println(relPath)
	}

	return nil
}

func (command FilesCommand) listFiles(args []string, explicitOnly bool) error {
	if len(args) == 0 {
		return errors.New("At least one tag must be specified. Use --all to show all files.")
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

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

		tag, err := store.TagByName(tagName)
		if err != nil {
			return err
		}
		if tag == nil {
			log.Fatalf("no such tag '%v'.", tagName)
		}

		if include {
			includeTagIds = append(includeTagIds, tag.Id)
		} else {
			excludeTagIds = append(excludeTagIds, tag.Id)
		}
	}

	if command.verbose {
		log.Info("retrieving set of tagged files from the database.")
	}

	var files database.Files
	if explicitOnly {
		files, err = store.FilesWithExplicitTags(includeTagIds, excludeTagIds)
		if err != nil {
			return err
		}
	} else {
		files, err = store.FilesWithTags(includeTagIds, excludeTagIds)
		if err != nil {
			return err
		}
	}

	for _, file := range files {
		relPath := common.RelPath(file.Path())
		fmt.Println(relPath)
	}

	return nil
}
