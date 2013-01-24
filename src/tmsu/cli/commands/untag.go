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
	"path/filepath"
	"strings"
	"tmsu/cli"
	"tmsu/log"
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
	return cli.Options{{"--all", "-a", "strip each file of all tags"},
		{"--tags", "-t", "the set of tags to remove"}}
}

func (command UntagCommand) Exec(options cli.Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no arguments specified.")
	}

	command.verbose = options.HasOption("--verbose")

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if options.HasOption("--all") {
		if len(args) < 1 {
			return fmt.Errorf("files to untag must be specified.")
		}

		paths := args

		if err := command.untagPathsAll(store, paths); err != nil {
			return err
		}
	} else if options.HasOption("--tags") {
		if len(args) < 2 {
			return fmt.Errorf("quoted set of tags and at least one file to untag must be specified.")
		}

		tagNames := strings.Fields(args[0])
		paths := args[1:]

		if err := command.untagPaths(store, paths, tagNames); err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("tag to remove and files to untag must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		if err := command.untagPath(store, path, tagNames); err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPathsAll(store *storage.Storage, paths []string) error {
	for _, path := range paths {
		if err := command.untagPathAll(store, path); err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPaths(store *storage.Storage, paths []string, tagNames []string) error {
	for _, path := range paths {
		if err := command.untagPath(store, path, tagNames); err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) untagPathAll(store *storage.Storage, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("'%v': could not retrieve file: %v", path, err)
	}
	if file == nil {
		return fmt.Errorf("'%v': file is not tagged.")
	}

	if command.verbose {
		log.Infof("'%v': identifying tags applied.", absPath)
	}

	explicitTags, err := store.ExplicitTagsByFileId(file.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not get explicit taggings: %v", path, err)
	}

	err = command.removeImplicitFileTags(store, absPath, explicitTags)
	if err != nil {
		return err
	}

	if command.verbose {
		log.Infof("'%v': removing all tags.", absPath)
	}

	err = store.RemoveExplicitFileTagsByFileId(file.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not remove explicit taggings: %v", path, err)
	}

	err = command.removeUntaggedFile(store, file)
	if err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) untagPath(store *storage.Storage, path string, tagNames []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
	}

	tags, err := command.lookupTags(store, tagNames)
	if err != nil {
		return err
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("'%v': could not retrieve file: %v", path, err)
	}
	if file == nil {
		return fmt.Errorf("'%v': file is not tagged.")
	}

	if err := command.removeImplicitFileTags(store, absPath, tags); err != nil {
		return err
	}

	if err := command.untagFile(store, file, tags); err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) untagFile(store *storage.Storage, file *database.File, tags database.Tags) error {
	for _, tag := range tags {
		if command.verbose {
			log.Infof("'%v': unapplying tag '%v'.", file.Path(), tag.Name)
		}

		if err := command.removeExplicitTag(store, file, tag); err != nil {
			return err
		}
	}

	if err := command.removeUntaggedFile(store, file); err != nil {
		return err
	}

	return nil
}

func (command UntagCommand) removeUntaggedFile(store *storage.Storage, file *database.File) error {
	filetagCount, err := store.FileTagCountByFileId(file.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not get tag count: %v", file.Path(), err)
	}

	if filetagCount == 0 {
		if command.verbose {
			log.Infof("'%v': removing untagged file.", file.Path())
		}

		err = store.RemoveFile(file.Id)
		if err != nil {
			return fmt.Errorf("'%v': could not remove file: %v", file.Path(), err)
		}
	}

	return nil
}

func (command UntagCommand) removeImplicitFileTags(store *storage.Storage, path string, tags database.Tags) error {
	descendents, err := store.FilesByDirectory(path)
	if err != nil {
		return fmt.Errorf("'%v': could not get files for directory: %v", path, err)
	}

	if command.verbose && len(descendents) > 0 {
		log.Infof("'%v': removing implicit taggings.", path)
	}

	for _, descendent := range descendents {
		if err = command.removeImplicitFileTag(store, descendent, tags); err != nil {
			return err
		}
	}

	return nil
}

func (command UntagCommand) removeImplicitFileTag(store *storage.Storage, file *database.File, tags database.Tags) error {
	for _, tag := range tags {
		if command.verbose {
			log.Infof("'%v': removing implicit tag '%v'.", file.Path(), tag.Name)
		}

		if err := command.removeImplicitTag(store, file, tag); err != nil {
			return err
		}
	}

	return command.removeUntaggedFile(store, file)
}

func (UntagCommand) removeExplicitTag(store *storage.Storage, file *database.File, tag *database.Tag) error {
	exists, err := store.ExplicitFileTagExists(file.Id, tag.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not determine whether file is tagged '%v': %v", file.Path(), tag.Name, err)
	}
	if !exists {
		return fmt.Errorf("'%v': file is not tagged '%v'.", file.Path(), tag.Name)
	}

	if err = store.RemoveExplicitFileTag(file.Id, tag.Id); err != nil {
		return fmt.Errorf("'%v': could not remove explicit file tagging '%v': %v", file.Path(), tag.Name)
	}

	return nil
}

func (UntagCommand) removeImplicitTag(store *storage.Storage, file *database.File, tag *database.Tag) error {
	exists, err := store.ImplicitFileTagExists(file.Id, tag.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not determine whether file is tagged '%v': %v", file.Path(), tag.Name, err)
	}
	if !exists {
		return fmt.Errorf("'%v': file is not tagged '%v'.", tag.Name)
	}

	if err = store.RemoveImplicitFileTag(file.Id, tag.Id); err != nil {
		return fmt.Errorf("'%v': could not remove implicit tagging '%v': %v", file.Path(), tag.Name)
	}

	return nil
}

func (UntagCommand) lookupTags(store *storage.Storage, tagNames []string) (database.Tags, error) {
	tags := make(database.Tags, len(tagNames))

	for index, tagName := range tagNames {
		tag, err := store.TagByName(tagName)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
		}
		if tag == nil {
			return nil, fmt.Errorf("no such tag '%v'.", tagName)
		}

		tags[index] = tag
	}

	return tags, nil
}
