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

var ImplyCommand = &Command{
	Name:     "imply",
	Synopsis: "Creates a tag implication",
	Description: `tmsu [OPTION] imply TAG1 TAG2
tmsu imply --list

Creates a tag implication such that whenever TAG1 is applied, TAG2 is automatically applied.`,
	Options: Options{Option{"--delete", "-d", "deletes the tag implication", false, ""},
		Option{"--list", "-l", "lists the tag implications", false, ""}},
	Exec: implyExec,
}

func implyExec(options Options, args []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch {
	case options.HasOption("--list"):
		return listImplications(store)
	case options.HasOption("--delete"):
		if len(args) < 2 {
			return fmt.Errorf("Implying and implied tag must be specified.")
		}

		return deleteImplication(store, args[0], args[1])
	}

	if len(args) < 2 {
		return fmt.Errorf("Implying and implied tag must be specified.")
	}

	return addImplication(store, args[0], args[1])
}

// unexported

func listImplications(store *storage.Storage) error {
	log.Suppf("retrieving tag implications.")

	implications, err := store.Implications()
	if err != nil {
		return fmt.Errorf("could not retrieve implications: %v", err)
	}

	width := 0
	for _, implication := range implications {
		length := len(implication.ImplyingTag.Name)
		if length > width {
			width = length
		}
	}

	previousImplyingTagName := ""
	for _, implication := range implications {
		if implication.ImplyingTag.Name != previousImplyingTagName {
			if previousImplyingTagName != "" {
				fmt.Println()
			}

			previousImplyingTagName = implication.ImplyingTag.Name

			fmt.Printf("%*v -> %v", width, implication.ImplyingTag.Name, implication.ImpliedTag.Name)
		} else {
			fmt.Printf(", %v", implication.ImpliedTag.Name)
		}
	}
	fmt.Println()

	return nil
}

func addImplication(store *storage.Storage, tagName, impliedTagName string) error {
	tag, err := store.Db.TagByName(tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return fmt.Errorf("no such tag '%v'.", tagName)
	}

	impliedTag, err := store.Db.TagByName(impliedTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", impliedTagName, err)
	}
	if impliedTag == nil {
		return fmt.Errorf("no such tag '%v'.", impliedTagName)
	}

	log.Suppf("adding tag implication of '%v' to '%v'.", tagName, impliedTagName)

	if err = store.AddImplication(tag.Id, impliedTag.Id); err != nil {
		return fmt.Errorf("could not add tag implication of '%v' to '%v': %v", tagName, impliedTagName, err)
	}

	return nil
}

func deleteImplication(store *storage.Storage, tagName, impliedTagName string) error {
	tag, err := store.Db.TagByName(tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return fmt.Errorf("no such tag '%v'.", tagName)
	}

	impliedTag, err := store.Db.TagByName(impliedTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", impliedTagName, err)
	}
	if impliedTag == nil {
		return fmt.Errorf("no such tag '%v'.", impliedTagName)
	}

	log.Suppf("removing tag implication of '%v' to '%v'.", tagName, impliedTagName)

	if err = store.RemoveImplication(tag.Id, impliedTag.Id); err != nil {
		return fmt.Errorf("could not add delete tag implication of '%v' to '%v': %v", tagName, impliedTagName, err)
	}

	return nil
}
