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
	"path/filepath"
	"sort"
	"strconv"
	"tmsu/cli/ansi"
	"tmsu/cli/terminal"
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

When color is turned on, tags are shown in the following colors:

     White  An explicitly applied (regular) tag
      Cyan  Tag implied by other tag(s)
    Yellow  Tag is both explicitly applied and implied by other tag(s)

See the 'imply' subcommand for more information on implied tags.

Examples:

    $ tmsu tags
    tralala.mp3: mp3 music opera 
    $ tmsu tags tralala.mp3
    mp3  music  opera
    $ tmsu tags --count tralala.mp3
    3`,
	Options: Options{{"--all", "-a", "lists all of the tags defined", false, ""},
		{"--count", "-c", "lists the number of tags rather than their names", false, ""},
		{"", "-1", "list one tag per line", false, ""},
		{"--explicit", "-e", "do not show implied tags", false, ""}},
	Exec: tagsExec,
}

func tagsExec(options Options, args []string) error {
	showCount := options.HasOption("--count")
	onePerLine := options.HasOption("-1")
	explicitOnly := options.HasOption("--explicit")

	var colour bool
	if options.HasOption("--color") {
		when := options.Get("--color").Argument
		switch when {
		case "auto":
			colour = terminal.Colour() && terminal.Width() > 0
		case "":
		case "always":
			colour = true
		case "never":
			colour = false
		default:
			return fmt.Errorf("invalid argument '%v' for '--color'", when)
		}
	} else {
		colour = terminal.Colour() && terminal.Width() > 0
	}

	if options.HasOption("--all") {
		return listAllTags(showCount, onePerLine, colour)
	}

	return listTags(args, showCount, onePerLine, explicitOnly, colour)
}

func listAllTags(showCount, onePerLine, colour bool) error {
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

		tagNames := make(ansi.Strings, len(tags))
		for index, tag := range tags {
			tagNames[index] = ansi.String(tag.Name)
		}

		if onePerLine {
			renderSingleColumn(tagNames, 0)
		} else {
			renderColumns(tagNames, terminal.Width())
		}
	}

	return nil
}

func listTags(paths []string, showCount, onePerLine, explicitOnly, colour bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch len(paths) {
	case 0:
		return listTagsForWorkingDirectory(store, showCount, onePerLine, explicitOnly, colour)
	case 1:
		return listTagsForPath(store, paths[0], showCount, onePerLine, explicitOnly, colour)
	default:
		return listTagsForPaths(store, paths, showCount, onePerLine, explicitOnly, colour)
	}

	return nil
}

func listTagsForPath(store *storage.Storage, path string, showCount, onePerLine, explicitOnly, colour bool) error {
	log.Infof(2, "%v: retrieving tags.", path)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve absolute path: %v", path, err)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}

	var tagNames ansi.Strings
	if file != nil {
		tagNames, err = tagNamesForFile(store, file.Id, explicitOnly, colour)
		if err != nil {
			return err
		}
	} else {
		log.Infof(2, "%v: untagged", path)

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
		if onePerLine {
			renderSingleColumn(tagNames, 0)
		} else {
			renderColumns(tagNames, terminal.Width())
		}
	}

	return nil
}

func listTagsForPaths(store *storage.Storage, paths []string, showCount, onePerLine, explicitOnly, colour bool) error {
	wereErrors := false
	for _, path := range paths {
		log.Infof(2, "%v: retrieving tags.", path)

		file, err := store.FileByPath(path)
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		var tagNames ansi.Strings
		if file != nil {
			tagNames, err = tagNamesForFile(store, file.Id, explicitOnly, colour)
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
			if onePerLine {
				fmt.Println(path)
				renderSingleColumn(tagNames, 2)
			} else {
				fmt.Print(path + ": ")
				renderSingleLine(tagNames)
			}
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func listTagsForWorkingDirectory(store *storage.Storage, showCount, onePerLine, explicitOnly, colour bool) error {
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

		tagNames, err := tagNamesForFile(store, file.Id, explicitOnly, colour)
		if err != nil {
			return err
		}

		if showCount {
			fmt.Println(dirName + ": " + strconv.Itoa(len(tagNames)))
		} else {
			if onePerLine {
				fmt.Println(dirName)
				renderSingleColumn(tagNames, 2)
			} else {
				fmt.Print(dirName + ": ")
				renderSingleLine(tagNames)
			}
		}
	}

	return nil
}

func tagNamesForFile(store *storage.Storage, fileId entities.FileId, explicitOnly, colour bool) (ansi.Strings, error) {
	fileTags, err := store.FileTagsByFileId(fileId, explicitOnly)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve file-tags for file '%v': %v", fileId, err)
	}

	tagNames := make(ansi.Strings, len(fileTags))

	for index, fileTag := range fileTags {
		tag, err := store.Tag(fileTag.TagId)
		if err != nil {
			return nil, fmt.Errorf("could not lookup tag: %v", err)
		}
		if tag == nil {
			return nil, fmt.Errorf("tag '%v' does not exist", fileTag.TagId)
		}

		var tagName ansi.String
		if fileTag.ValueId == 0 {
			tagName = ansi.String(tag.Name)
		} else {
			value, err := store.Value(fileTag.ValueId)
			if err != nil {
				return nil, fmt.Errorf("could not lookup value: %v", err)
			}
			if value == nil {
				return nil, fmt.Errorf("value '%v' does not exist", fileTag.ValueId)
			}

			tagName = ansi.String(tag.Name + "=" + value.Name)
		}

		if colour {
			if fileTag.Implicit {
				if fileTag.Explicit {
					tagName = ansi.String(ansi.Yellow + tagName + ansi.Reset)
				} else {
					tagName = ansi.String(ansi.Cyan + tagName + ansi.Reset)
				}
			}
		}

		tagNames[index] = tagName
	}

	return tagNames, nil
}
