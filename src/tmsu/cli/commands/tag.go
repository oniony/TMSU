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
	"os"
	"path/filepath"
	"strings"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type TagCommand struct{}

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
	return cli.Options{{"-t", "--tags", "the set of tags to apply"}}
}

func (command TagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return errors.New("Too few arguments.")
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	if cli.HasOption(options, "--tags") {
		if len(args) < 2 {
			return errors.New("Quoted set of tags and at least one file to tag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		tags, err := command.lookupTags(store, tagNames)
		if err != nil {
			return err
		}

		err = command.tagPaths(store, paths, tags, true)
		if err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return errors.New("File to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		tags, err := command.lookupTags(store, tagNames)
		if err != nil {
			return err
		}

		err = command.tagPath(store, path, tags, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) lookupTags(store *storage.Storage, names []string) (database.Tags, error) {
	tags := make(database.Tags, len(names))

	for index, name := range names {
		err := cli.ValidateTagName(name)
		if err != nil {
			return nil, err
		}

		tag, err := store.TagByName(name)
		if err != nil {
			return nil, err
		}

		if tag == nil {
			log.Warnf("New tag '%v'.", name)
			tag, err = store.AddTag(name)
			if err != nil {
				return nil, err
			}
		}

		tags[index] = tag
	}

	return tags, nil
}

func (command TagCommand) tagPaths(store *storage.Storage, paths []string, tags database.Tags, explicit bool) error {
	for _, path := range paths {
		err := command.tagPath(store, path, tags, explicit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPath(store *storage.Storage, path string, tags database.Tags, explicit bool) error {
	osInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := cli.AddFile(store, absPath)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if explicit {
			_, err = store.AddExplicitFileTag(file.Id, tag.Id)
		} else {
			_, err = store.AddImplicitFileTag(file.Id, tag.Id)
		}

		if err != nil {
			return err
		}
	}

	if !osInfo.IsDir() {
		return nil
	}

	osFile, err := os.Open(path)
	if err != nil {
		return err
	}

	entryNames, err := osFile.Readdirnames(0)
	osFile.Close()
	if err != nil {
		return err
	}

	for _, entryName := range entryNames {
		entryPath := filepath.Join(path, entryName)
		command.tagPath(store, entryPath, tags, false)
	}

	return nil
}
