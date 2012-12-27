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
	"path/filepath"
	"strings"
	"tmsu/cli"
	"tmsu/storage"
	"tmsu/storage/database"
)

type UntagCommand struct {
	verbose bool
}

func (UntagCommand) Name() cli.CommandName {
	return "untag"
}

func (UntagCommand) Synopsis() string {
	return "Remove tags from files"
}

func (UntagCommand) Description() string {
	return `tmsu untag FILE TAG...
tmsu untag --all FILE...
tmsu untag --tags "TAG..." FILE...

Disassociates FILE with the TAGs specified.`
}

func (UntagCommand) Options() cli.Options {
	return cli.Options{{"-a", "--all", "strip each file of all tags"},
		{"-t", "--tags", "the set of tags to remove"}}
}

func (command UntagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return errors.New("No arguments specified.")
	}

	command.verbose = cli.HasOption(options, "--verbose")

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	if cli.HasOption(options, "--all") {
		if len(args) < 1 {
			return errors.New("Files to untag must be specified.")
		}

		paths := args

		err := command.untagPathsAll(store, paths)
		if err != nil {
			return err
		}
	} else if cli.HasOption(options, "--tags") {
		if len(args) < 2 {
			return errors.New("Quoted set of tags and at least one file to untag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		err := command.untagPaths(store, paths, tagNames)
		if err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return errors.New("Tag to remove and files to untag must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		err := command.untagPath(store, path, tagNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPathsAll(store *storage.Storage, paths []string) error {
	for _, path := range paths {
		err := command.untagPathAll(store, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPaths(store *storage.Storage, paths []string, tagNames []string) error {
	for _, path := range paths {
		err := command.untagPath(store, path, tagNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPathAll(store *storage.Storage, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if command.verbose {
		fmt.Printf("'%v': removing all taggings.\n", path)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("File '" + path + "' is not tagged.")
	}

	if command.verbose {
		fmt.Printf("'%v': identifying tags applied.\n", path)
	}

	filetags, err := store.FileTagsByFileId(file.Id, true)
	if err != nil {
		return err
	}

	tags := make(database.Tags, len(filetags))
	for index, filetag := range filetags {
		tag, err := store.Tag(filetag.TagId)
		if err != nil {
			return err
		}

		tags[index] = tag
	}

	if command.verbose {
		fmt.Printf("'%v': found %v tags applied.\n", path, len(tags))
	}

	err = command.untagDescendents(store, absPath, tags)
	if err != nil {
		return err
	}

	err = command.untagFile(store, file, tags)
	if err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) untagPath(store *storage.Storage, path string, tagNames []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	tags, err := command.lookupTags(store, tagNames)
	if err != nil {
		return err
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("File '" + path + "' is not tagged.")
	}

	err = command.untagDescendents(store, absPath, tags)
	if err != nil {
		return err
	}

	err = command.untagFile(store, file, tags)
	if err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) untagFile(store *storage.Storage, file *database.File, tags database.Tags) error {
	for _, tag := range tags {
		if command.verbose {
			fmt.Printf("'%v': unapplying tag '%v'.\n", file.Path(), tag.Name)
		}

		err := command.removeExplicitTag(store, file, tag)
		if err != nil {
			return err
		}
	}

	err := command.removeUntaggedFile(store, file)
	if err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) removeUntaggedFile(store *storage.Storage, file *database.File) error {
	filetagCount, err := store.FileTagCountByFileId(file.Id, false)
	if err != nil {
		return err
	}

	if filetagCount == 0 {
		if command.verbose {
			fmt.Printf("'%v': removing untagged file.\n", file.Path())
		}

		err := store.RemoveFile(file.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagDescendents(store *storage.Storage, path string, tags database.Tags) error {
	if command.verbose {
		fmt.Printf("'%v': removing implicit taggings from descendents.\n", path)
	}

	descendents, err := store.FilesByDirectory(path)
	if err != nil {
		return err
	}

	for _, descendent := range descendents {
		err = command.untagDescendent(store, descendent, tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagDescendent(store *storage.Storage, file *database.File, tags database.Tags) error {
	for _, tag := range tags {
		if command.verbose {
			fmt.Printf("'%v': unapplying tag '%v'.\n", file.Path(), tag.Name)
		}

		err := command.removeImplicitTag(store, file, tag)
		if err != nil {
			return err
		}
	}

	return command.removeUntaggedFile(store, file)
}

func (UntagCommand) removeExplicitTag(store *storage.Storage, file *database.File, tag *database.Tag) error {
	filetag, err := store.FileTagByFileIdAndTagId(file.Id, tag.Id)
	if err != nil {
		return err
	}
	if filetag == nil {
		return errors.New("File '" + file.Path() + "' is not tagged '" + tag.Name + "'.")
	}

	err = store.RemoveExplicitFileTag(filetag.Id)
	if err != nil {
		return err
	}

	return nil
}

func (UntagCommand) removeImplicitTag(store *storage.Storage, file *database.File, tag *database.Tag) error {
	filetag, err := store.FileTagByFileIdAndTagId(file.Id, tag.Id)
	if err != nil {
		return err
	}
	if filetag == nil {
		return errors.New("File '" + file.Path() + "' is not tagged '" + tag.Name + "'.")
	}

	err = store.RemoveImplicitFileTag(filetag.Id)
	if err != nil {
		return err
	}

	return nil
}

func (UntagCommand) lookupTags(store *storage.Storage, tagNames []string) (database.Tags, error) {
	tags := make(database.Tags, len(tagNames))

	for index, tagName := range tagNames {
		tag, err := store.TagByName(tagName)
		if err != nil {
			return nil, err
		}
		if tag == nil {
			return nil, errors.New("No such tag '" + tagName + "'.")
		}

		tags[index] = tag
	}

	return tags, nil
}
