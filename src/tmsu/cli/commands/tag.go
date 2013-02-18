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
	"strings"
	"time"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type TagCommand struct {
	verbose   bool
	recursive bool
}

func (TagCommand) Name() cli.CommandName {
	return "tag"
}

func (TagCommand) Synopsis() string {
	return "Apply tags to files"
}

func (TagCommand) Description() string {
	return `tmsu tag FILE TAG...
tmsu tag --tags "TAG..." FILE...

Tags the file FILE with the tag(s) specified.`
}

func (TagCommand) Options() cli.Options {
	return cli.Options{{"--tags", "-t", "the set of tags to apply"},
		{"--recursive", "-r", "recursively apply tags to directory contents"}}
}

func (command TagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Too few arguments.")
	}

	command.verbose = options.HasOption("--verbose")
	command.recursive = options.HasOption("--recursive")

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if options.HasOption("--tags") {
		if len(args) < 2 {
			return fmt.Errorf("quoted set of tags and at least one file to tag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		tagIds, err := command.lookupTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err := command.tagPaths(store, paths, tagIds); err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("File to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		tagIds, err := command.lookupTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err = command.tagPath(store, path, tagIds); err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) lookupTagIds(store *storage.Storage, names []string) ([]uint, error) {
	tags, err := store.TagsByNames(names)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve tags %v: %v", names, err)
	}

	for _, name := range names {
		if !tags.Any(func(tag *database.Tag) bool { return tag.Name == name }) {
			log.Infof("New tag '%v'.", name)

			tag, err := store.AddTag(name)
			if err != nil {
				return nil, fmt.Errorf("could not add tag '%v': %v", name, err)
			}

			tags = append(tags, tag)
		}
	}

	tagIds := make([]uint, len(tags))
	for index, tag := range tags {
		tagIds[index] = tag.Id
	}

	return tagIds, nil
}

func (command TagCommand) tagPaths(store *storage.Storage, paths []string, tagIds []uint) error {
	for _, path := range paths {
		if err := command.tagPath(store, path, tagIds); err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPath(store *storage.Storage, path string, tagIds []uint) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return fmt.Errorf("%v: permisison denied", path)
		case os.IsNotExist(err):
			return fmt.Errorf("%v: no such file", path)
		default:
			return fmt.Errorf("%v: could not stat file: %v", path, err)
		}
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		file, err = command.addFile(store, absPath, stat.ModTime(), uint(stat.Size()), stat.IsDir())
		if err != nil {
			return fmt.Errorf("%v: could not add file: %v", path, err)
		}
	}

	if command.verbose {
		log.Infof("%v: applying tags.", file.Path())
	}

	if err = store.AddFileTags(file.Id, tagIds); err != nil {
		return fmt.Errorf("%v: could not apply tags: %v", file.Path(), err)
	}

	if command.recursive && stat.IsDir() {
		if err = command.tagRecursively(store, path, tagIds); err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagRecursively(store *storage.Storage, path string, tagIds []uint) error {
	osFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%v: could not open path: %v", path, err)
	}

	childNames, err := osFile.Readdirnames(0)
	osFile.Close()
	if err != nil {
		return fmt.Errorf("%v: could not retrieve directory contents: %v", path, err)
	}

	for _, childName := range childNames {
		childPath := filepath.Join(path, childName)

		if err = command.tagPath(store, childPath, tagIds); err != nil {
			return err
		}
	}

	return nil
}

func (command *TagCommand) addFile(store *storage.Storage, path string, modTime time.Time, size uint, isDir bool) (*database.File, error) {
	if command.verbose {
		log.Infof("%v: adding file.", path)
	}

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, fmt.Errorf("%v: could not create fingerprint: %v", path, err)
	}

	file, err := store.AddFile(path, fingerprint, modTime, int64(size), isDir)
	if err != nil {
		return nil, fmt.Errorf("%v: could not add file to database: %v", path, err)
	}

	return file, nil
}
