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
	"github.com/oniony/TMSU/common/terminal"
	"github.com/oniony/TMSU/storage"
	"strings"
)

var ValuesCommand = Command{
	Name:        "values",
	Synopsis:    "List values",
	Usages:      []string{"tmsu values [OPTION]... [TAG]..."},
	Description: "Lists the values for TAGs. If no TAG is specified then all tags are listed.",
	Examples: []string{"$ tmsu values year\n2000\n2001\n2017",
		"$ tmsu values\n2000\n2001\n2017\ncheese\nopera",
		"$ tmsu values --count year\n3"},
	Options: Options{{"--count", "-c", "lists the number of values rather than their names", false, ""},
		{"", "-1", "list one value per line", false, ""}},
	Exec: valuesExec,
}

// unexported

func valuesExec(options Options, args []string, databasePath string) (error, warnings) {
	showCount := options.HasOption("--count")
	onePerLine := options.HasOption("-1")

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

	if len(args) == 0 {
		return listAllValues(store, tx, showCount, onePerLine), nil
	}

	return listValues(store, tx, args, showCount, onePerLine)
}

func listAllValues(store *storage.Storage, tx *storage.Tx, showCount, onePerLine bool) error {
	log.Info(2, "retrieving all values.")

	if showCount {
		count, err := store.ValueCount(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve value count: %v", err)
		}

		fmt.Println(count)
	} else {
		values, err := store.Values(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve values: %v", err)
		}

		if onePerLine {
			for _, value := range values {
				fmt.Println(escape(value.Name))
			}
		} else {
			valueNames := make([]string, len(values))
			for index, value := range values {
				valueNames[index] = escape(value.Name)
			}

			terminal.PrintColumns(valueNames)
		}
	}

	return nil
}

func listValues(store *storage.Storage, tx *storage.Tx, args []string, showCount, onePerLine bool) (error, warnings) {
	tagNames := make([]string, len(args))
	for index, arg := range args {
		tagNames[index] = parseTagOrValueName(arg)
	}

	switch len(tagNames) {
	case 0:
		return fmt.Errorf("at least one tag must be specified"), nil
	case 1:
		return listValuesForTag(store, tx, tagNames[0], showCount, onePerLine), nil
	default:
		return listValuesForTags(store, tx, tagNames, showCount, onePerLine)
	}
}

func listValuesForTag(store *storage.Storage, tx *storage.Tx, tagName string, showCount, onePerLine bool) error {
	tag, err := store.TagByName(tx, tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return fmt.Errorf("no such tag, '%v'", tagName)
	}

	log.Infof(2, "retrieving values for tag '%v'.", tagName)

	values, err := store.ValuesByTag(tx, tag.Id)
	if err != nil {
		return fmt.Errorf("could not retrieve values for tag '%v': %v", tagName, err)
	}

	if showCount {
		fmt.Println(len(values))
	} else {
		if onePerLine {
			for _, value := range values {
				fmt.Println(escape(value.Name, '=', ' '))
			}
		} else {
			valueNames := make([]string, len(values))
			for index, value := range values {
				valueNames[index] = escape(value.Name, '=', ' ')
			}

			terminal.PrintColumns(valueNames)
		}
	}

	return nil
}

func listValuesForTags(store *storage.Storage, tx *storage.Tx, tagNames []string, showCount, onePerLine bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, tagName := range tagNames {
		tag, err := store.TagByName(tx, tagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err), warnings
		}
		if tag == nil {
			warnings = append(warnings, fmt.Sprintf("no such tag, '%v'.", tagName))
			continue
		}

		log.Infof(2, "retrieving values for tag '%v'.", tagName)

		values, err := store.ValuesByTag(tx, tag.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve values for tag '%v': %v", tagName, err), warnings
		}

		if showCount {
			fmt.Printf("%v: %v\n", tagName, len(values))
		} else {
			if onePerLine {
				fmt.Println(tagName)
				for _, value := range values {
					fmt.Println(escape(value.Name, '=', ' '))
				}
				fmt.Println()
			} else {
				valueNames := make([]string, len(values))
				for index, value := range values {
					valueNames[index] = escape(value.Name, '=', ' ')
				}

				fmt.Printf("%v: %v\n", tagName, strings.Join(valueNames, " "))
			}
		}
	}

	return nil, warnings
}
