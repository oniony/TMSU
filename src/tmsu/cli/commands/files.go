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
	"os"
	"path/filepath"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/path"
	"tmsu/storage"
)

type FilesCommand struct {
	verbose   bool
	directory bool
	file      bool
	top       bool
	leaf      bool
	recursive bool
}

func (FilesCommand) Name() cli.CommandName {
	return "files"
}

func (FilesCommand) Synopsis() string {
	return "List files with particular tags"
}

func (FilesCommand) Description() string {
	return `tmsu files [OPTION]... TAG...

Lists the files, if any, that have all of the TAGs specified. Tags can be
excluded by prefixing their names with a minus character (option processing
must first be disabled with '--').`
}

func (FilesCommand) Options() cli.Options {
	return cli.Options{{"--all", "-a", "list the complete set of tagged files", false, ""},
		{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--top", "-t", "list only the top-most matching items (excludes the contents of matching directories)", false, ""},
		{"--leaf", "-l", "list only the bottom-most (leaf) items", false, ""},
		{"--recursive", "-r", "read all files on the file-system under each matching directory, recursively", false, ""}}
}

func (command FilesCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.directory = options.HasOption("--directory")
	command.file = options.HasOption("--file")
	command.top = options.HasOption("--top")
	command.leaf = options.HasOption("--leaf")
	command.recursive = options.HasOption("--recursive")

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

	tree := path.NewTree()
	for _, file := range files {
		if command.directory && !file.IsDir {
			continue
		}
		if command.file && file.IsDir {
			continue
		}

		tree.Add(file.Path())
	}

	if command.top {
		tree = tree.TopLevel()
	} else {
		if command.recursive {
			paths, err := command.filesystemFiles(tree.TopLevel().Paths())
			if err != nil {
				return err
			}

			for _, path := range paths {
				tree.Add(path)
			}
		}
	}

	if command.leaf {
		tree = tree.Leaves()
	}

	for _, absPath := range tree.Paths() {
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

	tree := path.NewTree()
	for _, file := range files {
		if command.directory && !file.IsDir {
			continue
		}
		if command.file && file.IsDir {
			continue
		}

		tree.Add(file.Path())
	}

	if command.top {
		tree = tree.TopLevel()
		if err != nil {
			return fmt.Errorf("could not find top-level items: %v", err)
		}
	} else {
		if command.recursive {
			paths, err := command.filesystemFiles(tree.TopLevel().Paths())
			if err != nil {
				return err
			}

			for _, path := range paths {
				tree.Add(path)
			}
		}
	}

	if command.leaf {
		tree = tree.Leaves()
	}

	for _, absPath := range tree.Paths() {
		relPath := path.Rel(absPath)
		log.Print(relPath)
	}

	return nil
}

func (command *FilesCommand) filesystemFiles(paths []string) ([]string, error) {
	resultPaths := make([]string, 0, 100)

	for _, path := range paths {
		var err error
		resultPaths, err = command.filesystemFilesRecursive(path, resultPaths)
		if err != nil {
			return nil, err
		}
	}

	return resultPaths, nil
}

func (command *FilesCommand) filesystemFilesRecursive(path string, paths []string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return paths, nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied", path)
			return paths, nil
		default:
			return nil, fmt.Errorf("%v: could not stat: %v", path, err)
		}
	}

	if command.directory && !stat.IsDir() || command.file && stat.IsDir() {
		return paths, nil
	}

	paths = append(paths, path)

	if stat.IsDir() {
		dir, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not open directory: %v", path, err)
		}

		names, err := dir.Readdirnames(0)
		dir.Close()
		if err != nil {
			return nil, fmt.Errorf("%v: could not read directory entries: %v", path, err)
		}

		for _, name := range names {
			childPath := filepath.Join(path, name)
			paths, err = command.filesystemFilesRecursive(childPath, paths)
			if err != nil {
				return nil, err
			}
		}
	}

	return paths, nil
}
