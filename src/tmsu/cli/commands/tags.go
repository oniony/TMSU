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
	"path/filepath"
	"sort"
	"tmsu/cli"
	"tmsu/database"
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
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	tags, err := db.Tags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTags(paths []string, explicitOnly bool) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	switch len(paths) {
	case 0:
		return command.listTagsForWorkingDirectory(db, explicitOnly)
	case 1:
		return command.listTagsForPath(db, paths[0], explicitOnly)
	default:
		return command.listTagsForPaths(db, paths, explicitOnly)
	}

	return nil
}

func (command TagsCommand) listTagsForPath(db *database.Database, path string, explicitOnly bool) error {
	tags, err := command.tagsForPath(db, path, explicitOnly)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTagsForPaths(db *database.Database, paths []string, explicitOnly bool) error {
	for _, path := range paths {
		tags, err := command.tagsForPath(db, path, explicitOnly)
		if err != nil {
			return err
		}

		fmt.Print(path + ":")
		for _, tag := range tags {
			fmt.Print(" " + tag.Name)
		}
		fmt.Print("\n")
	}

	return nil
}

func (command TagsCommand) listTagsForWorkingDirectory(db *database.Database, explicitOnly bool) error {
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
		tags, err := command.tagsForPath(db, dirName, explicitOnly)
		if err != nil {
			return err
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

func (TagsCommand) tagsForPath(db *database.Database, path string, explicitOnly bool) (database.Tags, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	tags := make(database.Tags, 0, 10)
	for absPath != "/" {
		file, err := db.FileByPath(absPath)
		if err != nil {
			return nil, err
		}

		if file != nil {
			moreTags, err := db.TagsByFileId(file.Id)
			if err != nil {
				return nil, err
			}
			tags = append(tags, moreTags...)
		}

		if explicitOnly {
			break
		}

		absPath = filepath.Dir(absPath)
	}

	sort.Sort(tags)
	tags = uniq(tags)

	return tags, nil
}

func uniq(tags database.Tags) database.Tags {
	uniqueTags := make(database.Tags, 0, len(tags))

	var previousTagName string = ""
	for _, tag := range tags {
		if tag.Name == previousTagName {
			continue
		}

		uniqueTags = append(uniqueTags, tag)
		previousTagName = tag.Name
	}

	return uniqueTags
}
