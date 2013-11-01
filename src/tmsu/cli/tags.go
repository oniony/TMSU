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

package cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"tmsu/entities"
	"tmsu/log"
	"tmsu/storage"
)

var TagsCommand = Command{
	Name:     "tags",
	Synopsis: "List tags",
	Description: `tmsu tags [OPTION]... [FILE]...

Lists the tags applied to FILEs.

When run with no arguments, tags for the current working directory are listed.`,
	Options: Options{{"--all", "-a", "lists all of the tags defined", false, ""},
		{"--count", "-c", "lists the number of tags rather than their names", false, ""}},
	Exec: tagsExec,
}

func tagsExec(options Options, args []string) error {
	showCount := options.HasOption("--count")

	if options.HasOption("--all") {
		return listAllTags(showCount)
	}

	return listTags(args, showCount)
}

func listAllTags(showCount bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "retrieving all tags.")

	if showCount {
		count, err := store.TagCount()
		if err != nil {
			return fmt.Errorf("could not retrieve tag count: %v", err)
		}

		fmt.Println(count)
	} else {
		tags, err := store.Tags()
		if err != nil {
			return fmt.Errorf("could not retrieve tags: %v", err)
		}

		for _, tag := range tags {
			fmt.Println(tag.Name)
		}
	}

	return nil
}

func listTags(paths []string, showCount bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch len(paths) {
	case 0:
		return listTagsForWorkingDirectory(store, showCount)
	case 1:
		return listTagsForPath(store, paths[0], showCount)
	default:
		return listTagsForPaths(store, paths, showCount)
	}

	return nil
}

func listTagsForPath(store *storage.Storage, path string, showCount bool) error {
	log.Infof(2, "%v: retrieving tags.", path)

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

	if showCount {
		fmt.Println(len(tags))
	} else {
		for _, tag := range tags {
			fmt.Println(tag.Name)
		}
	}

	return nil
}

func listTagsForPaths(store *storage.Storage, paths []string, showCount bool) error {
	for _, path := range paths {
		log.Infof(2, "%v: retrieving tags.", path)

		var tags, err = store.TagsForPath(path)
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		if showCount {
			fmt.Println(path + ": " + strconv.Itoa(len(tags)))
		} else {
			fmt.Println(path + ": " + formatTags(tags))
		}
	}

	return nil
}

func listTagsForWorkingDirectory(store *storage.Storage, showCount bool) error {
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
		log.Infof(2, "%v: retrieving tags.", dirName)

		var tags, err = store.TagsForPath(dirName)

		if err != nil {
			log.Warn(err.Error())
			continue
		}

		if len(tags) == 0 {
			continue
		}

		if showCount {
			fmt.Println(dirName + ": " + strconv.Itoa(len(tags)))
		} else {
			fmt.Println(dirName + ": " + formatTags(tags))
		}
	}

	return nil
}

func formatTags(tags entities.Tags) string {
	tagNames := make([]string, len(tags))
	for index, tag := range tags {
		tagNames[index] = tag.Name
	}

	return strings.Join(tagNames, " ")
}
