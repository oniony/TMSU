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
	"tmsu/log"
	"tmsu/storage"
)

var MergeCommand = Command{
	Name:     "merge",
	Synopsis: "Merge tags",
	Description: `tmsu merge TAG... DEST
        
Merges TAGs into tag DEST resulting in a single tag of name DEST.

Examples:

    $ tmsu merge cehese cheese
    $ tmsu merge outdoors outdoor outside`,
	Options: Options{},
	Exec:    mergeExec,
}

func mergeExec(options Options, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("too few arguments.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	destTagName := args[len(args)-1]
	destTag, err := store.TagByName(destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
	}
	if destTag == nil {
		return fmt.Errorf("no such tag '%v'.", destTagName)
	}

	for _, sourceTagName := range args[0 : len(args)-1] {
		if sourceTagName == destTagName {
			return fmt.Errorf("source and destination names are the same.")
		}

		sourceTag, err := store.TagByName(sourceTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
		}
		if sourceTag == nil {
			return fmt.Errorf("no such tag '%v'.", sourceTagName)
		}

		log.Infof(2, "finding files tagged '%v'.", sourceTagName)

		fileTags, err := store.FileTagsByTagId(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve files for tag '%v': %v", sourceTagName, err)
		}

		log.Infof(2, "applying tag '%v' to these files.", destTagName)

		for _, fileTag := range fileTags {
			_, err = store.AddFileTag(fileTag.FileId, destTag.Id, fileTag.ValueId)
			if err != nil {
				return fmt.Errorf("could not apply tag '%v' to file #%v: %v", destTagName, fileTag.FileId, err)
			}
		}

		log.Infof(2, "untagging files '%v'.", sourceTagName)

		if err := store.RemoveFileTagsByTagId(sourceTag.Id); err != nil {
			return fmt.Errorf("could not remove all applications of tag '%v': %v", sourceTagName, err)
		}

		log.Infof(2, "updating tag implications involving tag '%v'.", sourceTagName)

		if err := store.UpdateImplicationsForTagId(sourceTag.Id, destTag.Id); err != nil {
			return fmt.Errorf("could not update tag implications involving tag '%v': %v", sourceTagName, err)
		}

		log.Infof(2, "deleting tag '%v'.", sourceTagName)

		err = store.DeleteTag(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", sourceTagName, err)
		}
	}

	return nil
}
