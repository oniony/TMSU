/*
Copyright 2011-2014 Paul Ruane.

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
	"tmsu/common/log"
	"tmsu/entities"
	"tmsu/storage"
)

var TagsCommand = Command{
	Name:     "tags",
	Synopsis: "List tags",
	Description: `tmsu tags [OPTION]... [FILE]...

Lists the tags applied to FILEs.

When run with no arguments, tags for the current working directory are listed.

Examples:

    $ tmsu tags
    tralala.mp3: mp3 music opera 
    $ tmsu tags tralala.mp3
    mp3
    music
    opera
    $ tmsu tags --count tralala.mp3
    3`,
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

	file, err := store.FileByPath(path)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}

	var tagNames []string
	if file != nil {
		fileTags, err := store.FileTagsByFileId(file.Id)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file-tags: %v", path, err)
		}

		tagNames, err = lookupTagNames(store, fileTags)
		if err != nil {
			return err
		}
	} else {
		_, err := os.Stat(path)
		if err != nil {
			switch {
			case os.IsPermission(err):
				return fmt.Errorf("%v: permission denied", path)
			case os.IsNotExist(err):
				return fmt.Errorf("%v: no such file", path)
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err)
			}
		}
	}

	if showCount {
		fmt.Println(len(tagNames))
	} else {
		for _, tagName := range tagNames {
			fmt.Println(tagName)
		}
	}

	return nil
}

func listTagsForPaths(store *storage.Storage, paths []string, showCount bool) error {
	wereErrors := false
	for _, path := range paths {
		log.Infof(2, "%v: retrieving tags.", path)

		file, err := store.FileByPath(path)
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		var tagNames []string
		if file != nil {
			fileTags, err := store.FileTagsByFileId(file.Id)
			if err != nil {
				return err
			}

			tagNames, err = lookupTagNames(store, fileTags)
			if err != nil {
				return err
			}
		} else {
			_, err := os.Stat(path)
			if err != nil {
				switch {
				case os.IsPermission(err):
					log.Warnf("%v: permission denied", path)
					wereErrors = true
					continue
				case os.IsNotExist(err):
					log.Warnf("%v: no such file", path)
					wereErrors = true
					continue
				default:
					return fmt.Errorf("%v: could not stat file: %v", path, err)
				}
			}
		}

		if showCount {
			fmt.Println(path + ": " + strconv.Itoa(len(tagNames)))
		} else {
			fmt.Println(path + ": " + strings.Join(tagNames, " "))
		}
	}

	if wereErrors {
		return blankError
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

		file, err := store.FileByPath(dirName)
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		if file == nil {
			continue
		}

		fileTags, err := store.FileTagsByFileId(file.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve file-tags: %v", err)
		}

		tagNames, err := lookupTagNames(store, fileTags)
		if err != nil {
			return err
		}

		if showCount {
			fmt.Println(dirName + ": " + strconv.Itoa(len(tagNames)))
		} else {
			fmt.Println(dirName + ": " + strings.Join(tagNames, " "))
		}
	}

	return nil
}

func lookupTagNames(store *storage.Storage, fileTags entities.FileTags) ([]string, error) {
	tagNames := make([]string, 0, len(fileTags))

	for _, fileTag := range fileTags {
		tag, err := store.Tag(fileTag.TagId)
		if err != nil {
			return nil, fmt.Errorf("could not lookup tag: %v", err)
		}
		if tag == nil {
			return nil, fmt.Errorf("tag '%v' does not exist", fileTag.TagId)
		}

		var tagName string
		if fileTag.ValueId == 0 {
			tagName = tag.Name
		} else {
			value, err := store.Value(fileTag.ValueId)
			if err != nil {
				return nil, fmt.Errorf("could not lookup value: %v", err)
			}
			if value == nil {
				return nil, fmt.Errorf("value '%v' does not exist", fileTag.ValueId)
			}

			tagName = tag.Name + "=" + value.Name
		}

		tagNames = append(tagNames, tagName)
	}

	sort.Strings(tagNames)

	return tagNames, nil
}
