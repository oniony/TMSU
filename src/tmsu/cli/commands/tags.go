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
	"sort"
	"strings"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type TagsCommand struct {
	verbose bool
}

func (TagsCommand) Name() cli.CommandName {
	return "tags"
}

func (TagsCommand) Synopsis() string {
	return "List tags"
}

func (TagsCommand) Description() string {
	return `tmsu tags [OPTION] [FILE]...
tmsu tags --all

Lists the tags applied to FILEs.

When run with no arguments, tags for the current working directory are listed.`
}

func (TagsCommand) Options() cli.Options {
	return cli.Options{{"--all", "-a", "lists all of the tags defined"}}
}

func (command TagsCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	if options.HasOption("--all") {
		return command.listAllTags()
	}

	return command.listTags(args)
}

func (command TagsCommand) listAllTags() error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.verbose {
		log.Info("retrieving all tags.")
	}

	tags, err := store.Tags()
	if err != nil {
		return fmt.Errorf("could not retrieve tags: %v", err)
	}

	for _, tag := range tags {
		log.Print(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTags(paths []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch len(paths) {
	case 0:
		return command.listTagsForWorkingDirectory(store)
	case 1:
		return command.listTagsForPath(store, paths[0])
	default:
		return command.listTagsForPaths(store, paths)
	}

	return nil
}

func (command TagsCommand) listTagsForPath(store *storage.Storage, path string) error {
	if command.verbose {
		log.Infof("%v: retrieving tags.", path)
	}

	var tags, err = store.TagsForPath(path)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve tags: %v", path, err)
	}

	if len(tags) == 0 {
		_, err := os.Stat(path)
		if err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("%v: permission denied", path)
			case os.IsNotExist(err):
				return fmt.Errorf("%v: file not found", path)
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err)
			}
		}
	}

	for _, tag := range tags {
		log.Print(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTagsForPaths(store *storage.Storage, paths []string) error {
	for _, path := range paths {
		if command.verbose {
			log.Infof("%v: retrieving tags.", path)
		}

		var tags, err = store.TagsForPath(path)

		if err != nil {
			log.Warn(err.Error())
			continue
		}

		log.Print(path + ": " + tagLine(tags))
	}

	return nil
}

func (command TagsCommand) listTagsForWorkingDirectory(store *storage.Storage) error {
	file, err := os.Open(".")
	if err != nil {
		return fmt.Errorf("could not open working directory: %v", err)
	}
	defer file.Close()

	dirNames, err := file.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("could not list working directory contents: %v", err)
	}

	sort.Strings(dirNames)

	for _, dirName := range dirNames {
		if command.verbose {
			log.Infof("%v: retrieving tags.", dirName)
		}

		var tags, err = store.TagsForPath(dirName)

		if err != nil {
			log.Warn(err.Error())
			continue
		}

		if len(tags) == 0 {
			continue
		}

		log.Print(dirName + ": " + tagLine(tags))
	}

	return nil
}

func tagLine(tags database.Tags) string {
	tagNames := make([]string, len(tags))
	for index, tag := range tags {
		tagNames[index] = tag.Name
	}

	return strings.Join(tagNames, " ")
}
