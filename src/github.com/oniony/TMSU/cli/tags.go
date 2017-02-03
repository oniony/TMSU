// Copyright 2011-2017 Paul Ruane.

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
	"github.com/oniony/TMSU/common/log"
	_path "github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/common/terminal"
	"github.com/oniony/TMSU/common/terminal/ansi"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
	"strconv"
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
		"$ tmsu tags --count tralala.mp3",
		"$ tmsu tags --value 2009 red"},
	Options: Options{{"--count", "-c", "lists the number of tags rather than their names", false, ""},
		{"", "-1", "list one tag per line", false, ""},
		{"--explicit", "-e", "do not show implied tags", false, ""},
		{"--name", "-n", "always print the file/value name", false, ""},
		{"--no-dereference", "-P", "do not follow symlinks (show tags for symlink itself)", false, ""},
		{"--value", "-u", "show tags which utilise values", false, ""}},
	Exec: tagsExec,
}

// unexported

func tagsExec(options Options, args []string, databasePath string) (error, warnings) {
	showCount := options.HasOption("--count")
	onePerLine := options.HasOption("-1")
	explicitOnly := options.HasOption("--explicit")
	printName := options.HasOption("--name")
	followSymlinks := !options.HasOption("--no-dereference")
	colour, err := useColour(options)
	if err != nil {
		return err, nil
	}

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	if options.HasOption("--value") {
		return listTagsForValues(store, tx, args, showCount, onePerLine, printName, colour)
	}

	if len(args) == 0 {
		return listAllTags(store, tx, showCount, onePerLine), nil
	}

	return listTagsForPaths(store, tx, args, showCount, onePerLine, explicitOnly, printName, colour, followSymlinks)
}

func listAllTags(store *storage.Storage, tx *storage.Tx, showCount, onePerLine bool) error {
	log.Info(2, "retrieving all tags.")

	if showCount {
		count, err := store.TagCount(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve tag count: %v", err)
		}

		fmt.Println(count)
	} else {
		tags, err := store.Tags(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve tags: %v", err)
		}

		if onePerLine {
			for _, tag := range tags {
				fmt.Println(escape(tag.Name, '=', ' '))
			}
		} else {
			tagNames := make([]string, len(tags))
			for index, tag := range tags {
				tagNames[index] = escape(tag.Name, '=', ' ')
			}

			terminal.PrintColumns(tagNames)
		}
	}

	return nil
}

func listTagsForPaths(store *storage.Storage, tx *storage.Tx, paths []string, showCount, onePerLine, explicitOnly, printPath, colour, followSymlinks bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	printPath = printPath || len(paths) > 1 || !stdoutIsCharDevice()

	for index, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err, warnings
		}

		log.Infof(2, "%v: resolving path", absPath)

		stat, err := os.Lstat(absPath)
		if err != nil {
			switch {
			case os.IsNotExist(err), os.IsPermission(err):
				stat = emptyStat{}
			default:
				warnings = append(warnings, err.Error())
				continue
			}
		} else if stat.Mode()&os.ModeSymlink != 0 && followSymlinks {
			absPath, err = _path.Dereference(absPath)
			if err != nil {
				warnings = append(warnings, err.Error())
				continue
			}
		}

		log.Infof(2, "%v: retrieving tags", absPath)

		file, err := store.FileByPath(tx, absPath)
		if err != nil {
			warnings = append(warnings, err.Error())
			continue
		}

		var tagNames []string
		if file != nil {
			tagNames, err = tagNamesForFile(store, tx, file.Id, explicitOnly, colour)
			if err != nil {
				return err, warnings
			}
		} else {
			_, err := os.Stat(absPath)
			if err != nil {
				switch {
				case os.IsPermission(err):
					warnings = append(warnings, fmt.Sprintf("%v: permission denied", absPath))
					continue
				case os.IsNotExist(err):
					warnings = append(warnings, fmt.Sprintf("%v: no such file", absPath))
					continue
				default:
					return fmt.Errorf("%v: could not stat file: %v", absPath, err), warnings
				}
			}
		}

		escapedPath := escape(path, '\\', ':')
		switch {
		case showCount:
			if printPath {
				fmt.Print(escapedPath + ": ")
			}

			fmt.Println(strconv.Itoa(len(tagNames)))
		case onePerLine:
			if index > 0 {
				fmt.Println()
			}

			if printPath {
				fmt.Println(escapedPath + ":")
			}

			for _, tagName := range tagNames {
				fmt.Println(tagName)
			}
		default:
			if printPath {
				fmt.Print(escapedPath + ":")

				for _, tagName := range tagNames {
					fmt.Print(" " + tagName)
				}

				fmt.Println()
			} else {
				terminal.PrintColumns(tagNames)
			}
		}
	}

	return nil, warnings
}

func listTagsForValues(store *storage.Storage, tx *storage.Tx, valueNames []string, showCount, onePerLine, printTagName, colour bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	printTagName = printTagName || len(valueNames) > 1 || !stdoutIsCharDevice()

	for index, valueName := range valueNames {
		log.Infof(2, "%v: looking up value", valueName)

		value, err := store.ValueByName(tx, valueName)
		if err != nil {
			return err, warnings
		}
		if value == nil {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", valueName))
			continue
		}

		log.Infof(2, "%v: retrieving tags", valueName)

		var tagNames []string
		if value != nil {
			tagNames, err = tagNamesForValue(store, tx, value.Id)
			if err != nil {
				return err, warnings
			}
		} else {
			warnings = append(warnings, fmt.Sprintf("value '%v' does not exist", valueName))
			continue
		}

		switch {
		case showCount:
			if printTagName {
				fmt.Println(valueName + ":")
			}

			fmt.Println(strconv.Itoa(len(tagNames)))
		case onePerLine:
			if index > 0 {
				fmt.Println()
			}

			if printTagName {
				fmt.Println(valueName + ":")
			}

			for _, tagName := range tagNames {
				fmt.Println(tagName)
			}
		default:
			if printTagName {
				fmt.Print(valueName + ":")

				for _, tagName := range tagNames {
					fmt.Print(" " + tagName)
				}

				fmt.Println()
			} else {
				terminal.PrintColumns(tagNames)
			}
		}
	}

	return nil, warnings
}

func tagNamesForFile(store *storage.Storage, tx *storage.Tx, fileId entities.FileId, explicitOnly, colour bool) ([]string, error) {
	fileTags, err := store.FileTagsByFileId(tx, fileId, explicitOnly)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve file-tags for file '%v': %v", fileId, err)
	}

	taggings := make([]string, len(fileTags))

	for index, fileTag := range fileTags {
		tag, err := store.Tag(tx, fileTag.TagId)
		if err != nil {
			return nil, fmt.Errorf("could not lookup tag: %v", err)
		}
		if tag == nil {
			return nil, fmt.Errorf("tag '%v' does not exist", fileTag.TagId)
		}

		var tagging string
		if fileTag.ValueId == 0 {
			tagging = formatTagValueName(tag.Name, "", colour, fileTag.Implicit, fileTag.Explicit)
		} else {
			value, err := store.Value(tx, fileTag.ValueId)
			if err != nil {
				return nil, fmt.Errorf("could not lookup value: %v", err)
			}
			if value == nil {
				return nil, fmt.Errorf("value '%v' does not exist", fileTag.ValueId)
			}

			tagging = formatTagValueName(tag.Name, value.Name, colour, fileTag.Implicit, fileTag.Explicit)
		}

		taggings[index] = tagging
	}

	ansi.Sort(taggings)

	return taggings, nil
}

func tagNamesForValue(store *storage.Storage, tx *storage.Tx, valueId entities.ValueId) ([]string, error) {
	fileTags, err := store.FileTagsByValueId(tx, valueId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve file-tags for value '%v': %v", valueId, err)
	}

	tagNames := make([]string, 0, 10)

	for _, fileTag := range fileTags {
		tag, err := store.Tag(tx, fileTag.TagId)
		if err != nil {
			return nil, fmt.Errorf("could not lookup tag: %v", err)
		}
		if tag == nil {
			return nil, fmt.Errorf("tag '%v' does not exist", fileTag.TagId)
		}

		if !containsTagName(tagNames, tag.Name) {
			tagNames = append(tagNames, tag.Name)
		}
	}

	ansi.Sort(tagNames)

	return tagNames, nil
}

func containsTagName(tagNames []string, search string) bool {
	for _, tagName := range tagNames {
		if tagName == search {
			return true
		}
	}

	return false
}
