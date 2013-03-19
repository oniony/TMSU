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

package commands

import (
	"fmt"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type ImplyCommand struct {
	verbose bool
}

func (ImplyCommand) Name() cli.CommandName {
	return "imply"
}

func (ImplyCommand) Synopsis() string {
	return "Creates a tag implication"
}

func (ImplyCommand) Description() string {
	return `tmsu [OPTION] imply TAG1 TAG2
tmsu imply --all

Creates a tag implication such that whenever TAG1 is applied, TAG2 is automatically applied.`
}

func (ImplyCommand) Options() cli.Options {
	return cli.Options{cli.Option{"--delete", "-d", "deletes the tag implication", false, ""},
		cli.Option{"--all", "-a", "lists the tag implications", false, ""}}
}

func (command ImplyCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch {
	case options.HasOption("--all"):
		return command.listImplications(store)
	case options.HasOption("--delete"):
		if len(args) < 2 {
			fmt.Errorf("Implying and implied tag must be specified.")
		}

		return command.deleteImplication(store, args[0], args[1])
	}

	if len(args) < 2 {
		fmt.Errorf("Implying and implied tag must be specified.")
	}

	return command.addImplication(store, args[0], args[1])
}

// unexported

func (command ImplyCommand) listImplications(store *storage.Storage) error {
	if command.verbose {
		log.Infof("retrieving tag implications.")
	}

	implications, err := store.Implications()
	if err != nil {
		return fmt.Errorf("could not retrieve implications: %v", err)
	}

	for _, implication := range implications {
		fmt.Printf("'%v' => '%v'\n", implication.ImplyingTag.Name, implication.ImpliedTag.Name)
	}

	return nil
}

func (command ImplyCommand) addImplication(store *storage.Storage, tagName, impliedTagName string) error {
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

	if command.verbose {
		log.Infof("adding tag implication of '%v' to '%v'.", tagName, impliedTagName)
	}

	if err = store.AddImplication(tag.Id, impliedTag.Id); err != nil {
		return fmt.Errorf("could not add delete tag implication of '%v' to '%v': %v", tagName, impliedTagName, err)
	}

	return nil
}

func (command ImplyCommand) deleteImplication(store *storage.Storage, tagName, impliedTagName string) error {
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

	if command.verbose {
		log.Infof("removing tag implication of '%v' to '%v'.", tagName, impliedTagName)
	}

	if err = store.RemoveImplication(tag.Id, impliedTag.Id); err != nil {
		return fmt.Errorf("could not add delete tag implication of '%v' to '%v': %v", tagName, impliedTagName, err)
	}

	return nil
}
