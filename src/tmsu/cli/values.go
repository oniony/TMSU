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
	"strings"
	"tmsu/common/log"
	"tmsu/storage"
)

var ValuesCommand = Command{
	Name:     "values",
	Synopsis: "List values",
	Description: `tmsu values [OPTION]... [TAG]...

Lists the values for TAGs.

Examples:

    $ tmsu values year
    2000
    2001
    2014
    $ tmsu values --all
    2000
    2001
    2014
    cheese
    opera
    $ tmsu values --count year
    3`,
	Options: Options{{"--all", "-a", "lists all of the values used by any tag", false, ""},
		{"--count", "-c", "lists the number of values rather than their names", false, ""}},
	Exec: valuesExec,
}

func valuesExec(options Options, args []string) error {
	showCount := options.HasOption("--count")

	if options.HasOption("--all") {
		return listAllValues(showCount)
	}

	if len(args) == 0 {
		return fmt.Errorf("at least one tag must be specified")
	}

	return listValues(args, showCount)
}

func listAllValues(showCount bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "retrieving all values.")

	if showCount {
		count, err := store.ValueCount()
		if err != nil {
			return fmt.Errorf("could not retrieve value count: %v", err)
		}

		fmt.Println(count)
	} else {
		values, err := store.Values()
		if err != nil {
			return fmt.Errorf("could not retrieve values: %v", err)
		}

		for _, value := range values {
			fmt.Println(value.Name)
		}
	}

	return nil
}

func listValues(tagNames []string, showCount bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch len(tagNames) {
	case 0:
		//TODO what to do?
	case 1:
		return listValuesForTag(store, tagNames[0], showCount)
	default:
		return listValuesForTags(store, tagNames, showCount)
	}

	return nil
}

func listValuesForTag(store *storage.Storage, tagName string, showCount bool) error {
	tag, err := store.TagByName(tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return fmt.Errorf("no such tag, '%v'.", tagName)
	}

	log.Infof(2, "retrieving values for tag '%v'.", tagName)

	values, err := store.ValuesByTag(tag.Id)
	if err != nil {
		return fmt.Errorf("could not retrieve values for tag '%v': %v", tagName, err)
	}

	if showCount {
		fmt.Println(len(values))
	} else {
		for _, value := range values {
			fmt.Println(value.Name)
		}
	}

	return nil
}

func listValuesForTags(store *storage.Storage, tagNames []string, showCount bool) error {
	for _, tagName := range tagNames {
		tag, err := store.TagByName(tagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
		}
		if tag == nil {
			return fmt.Errorf("no such tag, '%v'.", tagName)
		}

		log.Infof(2, "retrieving values for tag '%v'.", tagName)

		values, err := store.ValuesByTag(tag.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve values for tag '%v': %v", tagName, err)
		}

		if showCount {
			fmt.Printf("%v: %v\n", tagName, len(values))
		} else {
			valueNames := make([]string, len(values))
			for index, value := range values {
				valueNames[index] = value.Name
			}

			fmt.Printf("%v: %v\n", tagName, strings.Join(valueNames, " "))
		}
	}

	return nil
}
