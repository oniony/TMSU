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
	"tmsu/common/log"
	"tmsu/storage"
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

func mergeExec(store *storage.Storage, options Options, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments")
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	sourceNames := args[0 : len(args)-1]
	destName := args[len(args)-1]

	if options.HasOption("--value") {
		err = mergeValues(store, tx, sourceNames, destName)
	} else {
		err = mergeTags(store, tx, sourceNames, destName)
	}

	return err
}

func mergeTags(store *storage.Storage, tx *storage.Tx, sourceTagNames []string, destTagName string) error {
	destTag, err := store.TagByName(tx, destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
	}
	if destTag == nil {
		return fmt.Errorf("no such tag '%v'", destTagName)
	}

	wereErrors := false
	for _, sourceTagName := range sourceTagNames {
		if sourceTagName == destTagName {
			log.Warnf("cannot merge tag '%v' into itself", sourceTagName)
			wereErrors = true
			continue
		}

		sourceTag, err := store.TagByName(tx, sourceTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
		}
		if sourceTag == nil {
			log.Warnf("no such tag '%v'", sourceTagName)
			wereErrors = true
			continue
		}

		log.Infof(2, "finding files tagged '%v'.", sourceTagName)

		fileTags, err := store.FileTagsByTagId(tx, sourceTag.Id, true)
		if err != nil {
			return fmt.Errorf("could not retrieve files for tag '%v': %v", sourceTagName, err)
		}

		log.Infof(2, "applying tag '%v' to these files.", destTagName)

		for _, fileTag := range fileTags {
			_, err = store.AddFileTag(tx, fileTag.FileId, destTag.Id, fileTag.ValueId)
			if err != nil {
				return fmt.Errorf("could not apply tag '%v' to file #%v: %v", destTagName, fileTag.FileId, err)
			}
		}

		log.Infof(2, "deleting tag '%v'.", sourceTagName)

		err = store.DeleteTag(tx, sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", sourceTagName, err)
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func mergeValues(store *storage.Storage, tx *storage.Tx, sourceValueNames []string, destValueName string) error {
	destValue, err := store.ValueByName(tx, destValueName)
	if err != nil {
		return fmt.Errorf("could not retrieve value '%v': %v", destValueName, err)
	}
	if destValue == nil {
		return fmt.Errorf("no such value '%v'", destValueName)
	}

	wereErrors := false
	for _, sourceValueName := range sourceValueNames {
		if sourceValueName == destValueName {
			log.Warnf("cannot merge value '%v' into itself", sourceValueName)
			wereErrors = true
			continue
		}

		sourceValue, err := store.ValueByName(tx, sourceValueName)
		if err != nil {
			return fmt.Errorf("could not retrieve value '%v': %v", sourceValueName, err)
		}
		if sourceValue == nil {
			log.Warnf("no such value '%v'", sourceValueName)
			wereErrors = true
			continue
		}

		log.Infof(2, "finding files tagged with value '%v'.", sourceValueName)

		fileTags, err := store.FileTagsByValueId(tx, sourceValue.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve files for value '%v': %v", sourceValueName, err)
		}

		log.Infof(2, "applying value '%v' to these files.", destValueName)

		for _, fileTag := range fileTags {
			_, err = store.AddFileTag(tx, fileTag.FileId, fileTag.TagId, destValue.Id)
			if err != nil {
				return fmt.Errorf("could not apply value '%v' to file #%v: %v", destValueName, fileTag.FileId, err)
			}
		}

		log.Infof(2, "deleting value '%v'.", sourceValueName)

		err = store.DeleteValue(tx, sourceValue.Id)
		if err != nil {
			return fmt.Errorf("could not delete value '%v': %v", sourceValueName, err)
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}
