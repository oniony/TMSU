// Copyright 2011-2015 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"tmsu/common/log"
	"tmsu/common/terminal"
	"tmsu/common/terminal/ansi"
	"tmsu/entities"
	"tmsu/storage"
)

var TagsCommand = Command{
	Name:     "tags",
	Synopsis: "List tags",
	Usages:   []string{"tmsu tags [OPTION]... [FILE]..."},
	Description: `Lists the tags applied to FILEs. If no FILE is specified then all tags in the database are listed.

When color is turned on, tags are shown in the following colors:

  Normal  An explicitly applied (regular) tag
  'Cyan'    Tag implied by other tags
  'Yellow'  Tag is both explicitly applied and implied by other tags

See the 'imply' subcommand for more information on implied tags.`,
	Examples: []string{"$ tmsu tags\nmp3  music  opera",
		"$ tmsu tags tralala.mp3\nmp3  music  opera",
		"$ tmsu tags tralala.mp3 boom.mp3\n./tralala.mp3: mp3 music opera\n./boom.mp3: mp3 music drum-n-bass",
		"$ tmsu tags --count tralala.mp3"},
	Options: Options{{"--count", "-c", "lists the number of tags rather than their names", false, ""},
		{"", "-1", "list one tag per line", false, ""},
		{"--explicit", "-e", "do not show implied tags", false, ""}},
	Exec: tagsExec,
}

func tagsExec(store *storage.Storage, options Options, args []string) error {
	showCount := options.HasOption("--count")
	onePerLine := options.HasOption("-1")
	explicitOnly := options.HasOption("--explicit")
	colour, err := useColour(options)
	if err != nil {
		return err
	}

	if err := store.Begin(); err != nil {
		return err
	}
	defer store.Commit()

	if len(args) == 0 {
		return listAllTags(store, showCount, onePerLine, colour)
	}

	return listTagsForPaths(store, args, showCount, onePerLine, explicitOnly, colour)
}

func listAllTags(store *storage.Storage, showCount, onePerLine, colour bool) error {
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

		tagNames := make([]string, len(tags))
		for index, tag := range tags {
			tagNames[index] = tag.Name
		}

		if onePerLine {
			for _, tagName := range tagNames {
				fmt.Println(tagName)
			}
		} else {
			terminal.PrintColumns(tagNames)
		}
	}

	return nil
}

func listTagsForPaths(store *storage.Storage, paths []string, showCount, onePerLine, explicitOnly, colour bool) error {
	wereErrors := false
	printPath := len(paths) > 1 || terminal.Width() == 0

	for index, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		log.Infof(2, "%v: retrieving tags.", path)

		file, err := store.FileByPath(absPath)
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		var tagNames []string
		if file != nil {
			tagNames, err = tagNamesForFile(store, file.Id, explicitOnly, colour)
			if err != nil {
				return err
			}
		} else {
			_, err := os.Stat(absPath)
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

		switch {
		case showCount:
			if printPath {
				fmt.Print(path + ": ")
			}

			fmt.Println(strconv.Itoa(len(tagNames)))
		case onePerLine:
			if index > 0 {
				fmt.Println()
			}

			if printPath {
				fmt.Println(path + ":")
			}

			for _, tagName := range tagNames {
				fmt.Println(tagName)
			}
		default:
			if printPath {
				fmt.Print(path + ":")

				for _, tagName := range tagNames {
					fmt.Print(" " + tagName)
				}

				fmt.Println()
			} else {
				terminal.PrintColumns(tagNames)
			}
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func tagNamesForFile(store *storage.Storage, fileId entities.FileId, explicitOnly, colour bool) ([]string, error) {
	fileTags, err := store.FileTagsByFileId(fileId, explicitOnly)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve file-tags for file '%v': %v", fileId, err)
	}

	tagNames := make([]string, len(fileTags))

	for index, fileTag := range fileTags {
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

		if colour {
			if fileTag.Implicit {
				if fileTag.Explicit {
					tagName = ansi.Yellow(tagName)
				} else {
					tagName = ansi.Cyan(tagName)
				}
			}
		}

		tagNames[index] = tagName
	}

	ansi.Sort(tagNames)

	return tagNames, nil
}
