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
	"github.com/oniony/TMSU/storage"
)

var MergeCommand = Command{
	Name:        "merge",
	Synopsis:    "Merge tags",
	Usages:      []string{"tmsu merge TAG... DEST"},
	Description: `Merges TAGs into tag DEST resulting in a single tag of name DEST.`,
	Examples: []string{`$ tmsu merge cehese cheese`,
		`$ tmsu merge outdoors outdoor outside`},
	Options: Options{Option{"--value", "", "merge values", false, ""}},
	Exec:    mergeExec,
}

// unexported

func mergeExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments"), nil
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

	sourceNames := make([]string, len(args)-1)
	for index, name := range args[:len(args)-1] {
		sourceNames[index] = parseTagOrValueName(name)
	}

	destName := parseTagOrValueName(args[len(args)-1])

	if options.HasOption("--value") {
		return mergeValues(store, tx, sourceNames, destName)
	}

	return mergeTags(store, tx, sourceNames, destName)
}

func mergeTags(store *storage.Storage, tx *storage.Tx, sourceTagNames []string, destTagName string) (error, warnings) {
	destTag, err := store.TagByName(tx, destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err), nil
	}
	if destTag == nil {
		return fmt.Errorf("no such tag '%v'", destTagName), nil
	}

	warnings := make(warnings, 0, 10)
	for _, sourceTagName := range sourceTagNames {
		if sourceTagName == destTagName {
			warnings = append(warnings, fmt.Sprintf("cannot merge tag '%v' into itself", sourceTagName))
			continue
		}

		sourceTag, err := store.TagByName(tx, sourceTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err), warnings
		}
		if sourceTag == nil {
			warnings = append(warnings, fmt.Sprintf("no such tag '%v'", sourceTagName))
			continue
		}

		log.Infof(2, "finding files tagged '%v'.", sourceTagName)

		fileTags, err := store.FileTagsByTagId(tx, sourceTag.Id, true)
		if err != nil {
			return fmt.Errorf("could not retrieve files for tag '%v': %v", sourceTagName, err), warnings
		}

		log.Infof(2, "applying tag '%v' to these files.", destTagName)

		for _, fileTag := range fileTags {
			if _, err = store.AddFileTag(tx, fileTag.FileId, destTag.Id, fileTag.ValueId); err != nil {
				return fmt.Errorf("could not apply tag '%v' to file #%v: %v", destTagName, fileTag.FileId, err), warnings
			}
		}

		log.Infof(2, "deleting tag '%v'.", sourceTagName)

		if err = store.DeleteTag(tx, sourceTag.Id); err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", sourceTagName, err), warnings
		}
	}

	return nil, warnings
}

func mergeValues(store *storage.Storage, tx *storage.Tx, sourceValueNames []string, destValueName string) (error, warnings) {
	destValue, err := store.ValueByName(tx, destValueName)
	if err != nil {
		return fmt.Errorf("could not retrieve value '%v': %v", destValueName, err), nil
	}
	if destValue == nil {
		return fmt.Errorf("no such value '%v'", destValueName), nil
	}

	warnings := make(warnings, 0, 10)

	for _, sourceValueName := range sourceValueNames {
		if sourceValueName == destValueName {
			warnings = append(warnings, fmt.Sprintf("cannot merge value '%v' into itself", sourceValueName))
			continue
		}

		sourceValue, err := store.ValueByName(tx, sourceValueName)
		if err != nil {
			return fmt.Errorf("could not retrieve value '%v': %v", sourceValueName, err), warnings
		}
		if sourceValue == nil {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", sourceValueName))
			continue
		}

		log.Infof(2, "finding files tagged with value '%v'.", sourceValueName)

		fileTags, err := store.FileTagsByValueId(tx, sourceValue.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve files for value '%v': %v", sourceValueName, err), warnings
		}

		log.Infof(2, "applying value '%v' to these files.", destValueName)

		for _, fileTag := range fileTags {
			if _, err = store.AddFileTag(tx, fileTag.FileId, fileTag.TagId, destValue.Id); err != nil {
				return fmt.Errorf("could not apply value '%v' to file #%v: %v", destValueName, fileTag.FileId, err), warnings
			}
		}

		log.Infof(2, "deleting value '%v'.", sourceValueName)

		if err = store.DeleteValue(tx, sourceValue.Id); err != nil {
			return fmt.Errorf("could not delete value '%v': %v", sourceValueName, err), warnings
		}
	}

	return nil, warnings
}
