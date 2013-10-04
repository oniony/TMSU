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
	"tmsu/storage/database"
)

type FilesCommand struct {
	verbose   bool
	directory bool
	file      bool
	top       bool
	leaf      bool
	recursive bool
	print0    bool
	count     bool
}

func (FilesCommand) Name() cli.CommandName {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "List files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files [OPTION]... QUERY 

Lists the files that match the QUERY specified. QUERY may contain tag names,
logical operators ('not', 'and', 'or') and parentheses.

If multiple tags are specified without a logical operator between them this
will be interpretted as an implicit 'and', e.g. 'chalk cheese' is interpretted as
'chalk and cheese'.

Examples:

    $ tmsu files music
    $ tmsu files music mp3                # implict 'and'
    $ tmsu files music and mp3            # same query but with explicit 'and'
    $ tmsu files music not flac
    $ tmsu files music and not flac       # same query with explicit 'and'
    $ tmsu files mp3 or flac
    $ tmsu files music mp3 or flac
    $ tmsu files (music and mp3) or flac  # same query written in full`
}

func (FilesCommand) Options() cli.Options {
	return cli.Options{{"--all", "-a", "list the complete set of tagged files", false, ""},
		{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--top", "-t", "list only the top-most matching items (excludes the contents of matching directories)", false, ""},
		{"--leaf", "-l", "list only the bottom-most (leaf) items", false, ""},
		{"--recursive", "-r", "read all files on the file-system under each matching directory, recursively", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""}}
}

func (command FilesCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.directory = options.HasOption("--directory")
	command.file = options.HasOption("--file")
	command.top = options.HasOption("--top")
	command.leaf = options.HasOption("--leaf")
	command.recursive = options.HasOption("--recursive")
	command.print0 = options.HasOption("--print0")
	command.count = options.HasOption("--count")

	if options.HasOption("--all") {
		return command.listAllFiles()
	}

	return command.listFilesForTags(args)
}

// unexported

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

	return command.listFiles(files)
}

func (command FilesCommand) listFilesForTags(args []string) error {
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

	return command.listFiles(files)
}

func (command *FilesCommand) listFiles(files database.Files) error {
	tree := path.NewTree()
	for _, file := range files {
		tree.Add(file.Path(), file.IsDir)
	}

	if command.top {
		tree = tree.TopLevel()
	} else {
		if command.recursive {
			fsFiles, err := path.Enumerate(tree.TopLevel().Paths())
			if err != nil {
				return err
			}

			for _, fsFile := range fsFiles {
				tree.Add(fsFile.Path, fsFile.IsDir)
			}
		}
	}

	if command.leaf {
		tree = tree.Leaves()
	}

	if command.file {
		tree = tree.Files()
	}

	if command.directory {
		tree = tree.Directories()
	}

	if command.count {
		log.Print(len(tree.Paths()))
	} else {
		for _, absPath := range tree.Paths() {
			relPath := path.Rel(absPath)

			if command.print0 {
				log.Print0(relPath)
			} else {
				log.Print(relPath)
			}
		}
	}

	return nil
}
