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
	"fmt"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/path"
	"tmsu/storage"
)

type FilesCommand struct {
	verbose   bool
	directory bool
	file      bool
	branch    bool
	leaf      bool
}

func (FilesCommand) Name() cli.CommandName {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "List files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files OPTIONS [-]TAG...

Lists the files, if any, that have all of the TAGs specified. Tags can be
excluded by prefixing their names with a minus character (option processing
must first be disabled with '--').`
}

func (FilesCommand) Options() cli.Options {
	return cli.Options{{"--all", "-a", "show the complete set of tagged files"},
		{"--directory", "-d", "list only items that are directories"},
		{"--file", "-f", "list only items that are files"},
		{"--branch", "-b", "list only matching branches (omit directory contents)"},
		{"--leaf", "-l", "list only leaf items (omit parent directories)"}}
}

func (command FilesCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.directory = options.HasOption("--directory")
	command.file = options.HasOption("--file")
	command.branch = options.HasOption("--branch")
	command.leaf = options.HasOption("--leaf")

	if options.HasOption("--all") {
		return command.listAllFiles()
	}

	return command.listFiles(args)
}

func (command FilesCommand) listAllFiles() error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.verbose {
		log.Info("retrieving all files from database.")
	}

	files, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve files: %v", err)
	}

	absPaths := make([]string, 0, len(files))
	for _, file := range files {
		if command.directory && !file.IsDir {
			continue
		}
		if command.file && file.IsDir {
			continue
		}

		absPaths = append(absPaths, file.Path())
	}

	if command.branch {
		absPaths, err = path.NonNested(absPaths)
		if err != nil {
			return fmt.Errorf("could not find branch entries: %v", err)
		}
	}

	if command.leaf {
		absPaths, err = path.Leaves(absPaths)
		if err != nil {
			return fmt.Errorf("could not find leaf entries: %v", err)
		}
	}

	for _, absPath := range absPaths {
		relPath := path.Rel(absPath)
		log.Print(relPath)
	}

	return nil
}

func (command FilesCommand) listFiles(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one tag must be specified. Use --all to show all files.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
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
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
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

	files, err := store.FilesWithTags(includeTagIds, excludeTagIds)
	if err != nil {
		return fmt.Errorf("could not retrieve files with tags %v and without tags %v: %v", includeTagIds, excludeTagIds, err)
	}

	absPaths := make([]string, 0, len(files))
	for _, file := range files {
		if command.directory && !file.IsDir {
			continue
		}
		if command.file && file.IsDir {
			continue
		}

		absPaths = append(absPaths, file.Path())
	}

	if command.branch {
		absPaths, err = path.NonNested(absPaths)
		if err != nil {
			return fmt.Errorf("could not find branch items: %v", err)
		}
	}

	if command.leaf {
		absPaths, err = path.Leaves(absPaths)
		if err != nil {
			return fmt.Errorf("could not find leaf items: %v", err)
		}
	}

	for _, absPath := range absPaths {
		relPath := path.Rel(absPath)
		log.Print(relPath)
	}

	return nil
}
