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
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type TagCommand struct {
	verbose bool
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
	return cli.Options{{"--tags", "-t", "the set of tags to apply"}}
}

func (command TagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Too few arguments.")
	}

	command.verbose = options.HasOption("--verbose")

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

		tags, err := command.lookupTags(store, tagNames)
		if err != nil {
			return err
		}

		files, err := command.addFiles(store, paths)
		if err != nil {
			return err
		}

		if err := command.tagFiles(store, files, tags, true); err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("File to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		tags, err := command.lookupTags(store, tagNames)
		if err != nil {
			return err
		}

		file, err := command.addFile(store, path)
		if err != nil {
			return err
		}

		if err = command.tagFile(store, file, tags, true); err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) lookupTags(store *storage.Storage, names []string) (database.Tags, error) {
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

	return tags, nil
}

func (command TagCommand) addFiles(store *storage.Storage, paths []string) (database.Files, error) {
	files := make(database.Files, len(paths))

	for index, path := range paths {
		file, err := command.addFile(store, path)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not add file: %v", path, err)
		}

		files[index] = file
	}

	return files, nil
}

func (command TagCommand) addFile(store *storage.Storage, path string) (*database.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not get absolute path: %v", path, err)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not retrieve file: %v", path, err)
	}

	if file != nil {
		return file, nil
	}

	if command.verbose {
		log.Infof("'%v': adding file.", path)
	}

	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return nil, fmt.Errorf("'%v': permisison denied", path)
		case os.IsNotExist(err):
			return nil, fmt.Errorf("'%v': no such file", path)
		default:
			return nil, fmt.Errorf("'%v': could not stat file: %v", path, err)
		}
	}

	modTime := stat.ModTime().UTC()
	size := stat.Size()

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not create fingerprint: %v", path, err)
	}

	file, err = store.AddFile(absPath, fingerprint, modTime, size)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not add file to database: %v", path, err)
	}

	return file, nil
}

func (command TagCommand) tagFiles(store *storage.Storage, files database.Files, tags database.Tags, explicit bool) error {
	for _, file := range files {
		if err := command.tagFile(store, file, tags, explicit); err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagFile(store *storage.Storage, file *database.File, tags database.Tags, explicit bool) error {
	tagIds := make([]uint, len(tags))
	for index, tag := range tags {
		tagIds[index] = tag.Id
	}

	if explicit {
		if command.verbose {
			log.Infof("'%v': applying explicit tags.", file.Path())
		}

		err := store.AddExplicitFileTags(file.Id, tagIds)
		if err != nil {
			return fmt.Errorf("'%v': could not apply explicit tags: %v", file.Path(), err)
		}
	} else {
		if command.verbose {
			log.Infof("'%v': applying implicit tags.", file.Path())
		}

		err := store.AddImplicitFileTags(file.Id, tagIds)
		if err != nil {
			return fmt.Errorf("'%v': could not apply implicit tags: %v", file.Path(), err)
		}
	}

	stat, err := os.Stat(file.Path())
	if err != nil {
		return fmt.Errorf("'%v': could not stat file: %v", file.Path(), err)
	}

	if !stat.IsDir() {
		return nil
	}

	osFile, err := os.Open(file.Path())
	if err != nil {
		return fmt.Errorf("'%v': could not open path: %v", file.Path(), err)
	}

	childNames, err := osFile.Readdirnames(0)
	osFile.Close()
	if err != nil {
		return fmt.Errorf("'%v': could not retrieve directory contents: %v", file.Path(), err)
	}

	for _, childName := range childNames {
		childPath := filepath.Join(file.Path(), childName)

		childFile, err := store.FileByPath(childPath)
		if err != nil {
			return fmt.Errorf("'%v': could not lookup file: %v", childPath, err)
		}
		if childFile == nil {
			childFile, err = command.addFile(store, childPath)
			if err != nil {
				return err
			}
		}

		err = command.tagFile(store, childFile, tags, false)
		if err != nil {
			return err
		}
	}

	return nil
}
