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
)

var CopyCommand = Command{
	Name:        "copy",
	Aliases:     []string{"cp"},
	Synopsis:    "Create a copy of a tag",
	Usages:      []string{"tmsu copy TAG NEW..."},
	Description: `Creates a new tag NEW applied to the same set of files as TAG.`,
	Examples: []string{"$ tmsu copy cheese wine",
		"$ tmsu copy report document text"},
	Options: Options{},
	Exec:    copyExec,
}

// unexported

func copyExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments"), nil
	}

	sourceTagName := parseTagOrValueName(args[0])

	destTagNames := make([]string, len(args)-1)
	for index, arg := range args[1:] {
		destTagNames[index] = parseTagOrValueName(arg)
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

	sourceTag, err := store.TagByName(tx, sourceTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err), nil
	}
	if sourceTag == nil {
		return fmt.Errorf("no such tag '%v'", sourceTagName), nil
	}

	warnings := make(warnings, 0, 10)

	for _, destTagName := range destTagNames {
		destTag, err := store.TagByName(tx, destTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err), warnings
		}
		if destTag != nil {
			warnings = append(warnings, fmt.Sprintf("a tag with name '%v' already exists", destTagName))
			continue
		}

		log.Infof(2, "copying tag '%v' to '%v'.", sourceTagName, destTagName)

		if _, err = store.CopyTag(tx, sourceTag.Id, destTagName); err != nil {
			return fmt.Errorf("could not copy tag '%v' to '%v': %v", sourceTagName, destTagName, err), warnings
		}
	}

	return nil, warnings
}
