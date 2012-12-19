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

	if cli.HasOption(options, "--tags") {
		if len(args) < 2 {
			return errors.New("Quoted set of tags and at least one file to tag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		err := command.tagPaths(paths, tagNames, true)
		if err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return errors.New("File to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		err := command.tagPath(path, tagNames, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPaths(paths []string, tagNames []string, explicit bool) error {
	for _, path := range paths {
		err := command.tagPath(path, tagNames, explicit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPath(path string, tagNames []string, explicit bool) error {
	osInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := cli.AddFile(store, absPath)
	if err != nil {
		return err
	}

	for _, tagName := range tagNames {
		_, _, err = command.applyTag(store, path, file.Id, tagName, explicit)
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
		command.tagPath(entryPath, tagNames, false)
	}

	return nil
}

func (TagCommand) applyTag(store *storage.Storage, path string, fileId uint, tagName string, explicit bool) (*database.Tag, *database.FileTag, error) {
	err := cli.ValidateTagName(tagName)
	if err != nil {
		return nil, nil, err
	}

	tag, err := store.TagByName(tagName)
	if err != nil {
		return nil, nil, err
	}

	if tag == nil {
		log.Warnf("New tag '%v'.", tagName)
		tag, err = store.AddTag(tagName)
		if err != nil {
			return nil, nil, err
		}
	}

	//TODO move this logic into Storage
	fileTag, err := store.FileTagByFileIdAndTagId(fileId, tag.Id)
	fmt.Println("Found filetag", fileTag)
	if err != nil {
		return nil, nil, err
	}

	if fileTag == nil {
		_, err := store.AddFileTag(fileId, tag.Id, explicit, !explicit)
		if err != nil {
			return nil, nil, err
		}
	} else {
		err := store.UpdateFileTag(fileTag.Id, fileTag.Explicit || explicit, fileTag.Implicit || !explicit)
		if err != nil {
			return nil, nil, err
		}
	}

	return tag, fileTag, nil
}
