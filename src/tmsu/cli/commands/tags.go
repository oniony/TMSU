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
	"fmt"
	"os"
	"sort"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type TagsCommand struct{}

func (TagsCommand) Name() string {
	return "tags"
}

func (TagsCommand) Synopsis() string {
	return "List tags"
}

func (TagsCommand) Description() string {
	return `tmsu tags [--explicit] [FILE]...
tmsu tags --all

Lists the tags applied to FILEs.

When run with no arguments, tags for the current working directory are listed.

  --all         lists all of the tags defined
  --explicit    show only explicitly applied tags (not inherited)`
}

func (TagsCommand) Options() []cli.Option {
	return []cli.Option{}
}

func (command TagsCommand) Exec(args []string) error {
	if len(args) == 1 && args[0] == "--all" {
		return command.listAllTags()
	}

	explicitOnly := false
	if len(args) > 0 && args[0] == "--explicit" {
		explicitOnly = true
		args = args[1:]
	}

	return command.listTags(args, explicitOnly)
}

func (TagsCommand) listAllTags() error {
	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	tags, err := store.Tags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTags(paths []string, explicitOnly bool) error {
	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	switch len(paths) {
	case 0:
		return command.listTagsForWorkingDirectory(store, explicitOnly)
	case 1:
		return command.listTagsForPath(store, paths[0], explicitOnly)
	default:
		return command.listTagsForPaths(store, paths, explicitOnly)
	}

	return nil
}

func (command TagsCommand) listTagsForPath(store *storage.Storage, path string, explicitOnly bool) error {
	tags, err := store.TagsForPath(path, explicitOnly)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTagsForPaths(store *storage.Storage, paths []string, explicitOnly bool) error {
	for _, path := range paths {
		tags, err := store.TagsForPath(path, explicitOnly)
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		fmt.Print(path + ":")
		for _, tag := range tags {
			fmt.Print(" " + tag.Name)
		}
		fmt.Print("\n")
	}

	return nil
}

func (command TagsCommand) listTagsForWorkingDirectory(store *storage.Storage, explicitOnly bool) error {
	file, err := os.Open(".")
	if err != nil {
		return err
	}
	defer file.Close()

	dirNames, err := file.Readdirnames(0)
	if err != nil {
		return err
	}

	sort.Strings(dirNames)

	for _, dirName := range dirNames {
		tags, err := store.TagsForPath(dirName, explicitOnly)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				// do nothing
			case os.IsPermission(err):
				log.Warnf("%v: Permission denied")
			default:
				return err
			}
		}

		if len(tags) == 0 {
			continue
		}

		fmt.Print(dirName + ":")
		for _, tag := range tags {
			fmt.Print(" " + tag.Name)
		}
		fmt.Print("\n")
	}

	return nil
}
