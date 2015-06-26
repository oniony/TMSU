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
	"tmsu/entities"
	"tmsu/storage"
)

var ImplyCommand = Command{
	Name:     "imply",
	Synopsis: "Creates a tag implication",
	Usages: []string{"tmsu imply [OPTION] TAG[=VALUE] IMPL[=VALUE]...",
		"tmsu imply"},
	Description: `Creates a tag implication such that any file tagged TAG will be implicitly tagged IMPL.

When run without arguments lists the set of tag implications.

Tag implications are applied at time of file query (not at time of tag application) therefore any changes to the implication rules will affect all further queries.

By default the 'tag' subcommand will not explicitly apply tags that are already implied by the implication rules.

The 'tags' subcommand can be used to identify which tags applied to a file are implied.`,
	Examples: []string{`$ tmsu imply mp3 music`,
		`$ tmsu imply
mp3 -> music`,
		`$ tmsu imply aubergine aka=eggplant`,
		`$ tmsu imply --delete mp3 music`},
	Options: Options{Option{"--delete", "-d", "deletes the tag implication", false, ""}},
	Exec:    implyExec,
}

func implyExec(store *storage.Storage, options Options, args []string) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	if options.HasOption("--delete") {
		if len(args) < 2 {
			return fmt.Errorf("too few arguments")
		}

		return deleteImplications(store, tx, args)
	}

	switch len(args) {
	case 0:
		return listImplications(store, tx)
	case 1:
		return fmt.Errorf("tag(s) to be implied must be specified")
	default:
		return addImplications(store, tx, args)
	}
}

// unexported

func listImplications(store *storage.Storage, tx *storage.Tx) error {
	log.Infof(2, "retrieving tag implications.")

	implications, err := store.Implications(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve implications: %v", err)
	}

	width := 0
	for _, implication := range implications {
		length := len(implication.ImplyingTag.Name)
		if implication.ImplyingValue.Id != 0 {
			length += 1 + len(implication.ImplyingValue.Name)
		}

		if length > width {
			width = length
		}
	}

	if len(implications) > 0 {
		for _, implication := range implications {
			implying := implication.ImplyingTag.Name
			if implication.ImplyingValue.Id != 0 {
				implying += "=" + implication.ImplyingValue.Name
			}

			implied := implication.ImpliedTag.Name
			if implication.ImpliedValue.Id != 0 {
				implied += "=" + implication.ImpliedValue.Name
			}

			fmt.Printf("%*v -> %v\n", width, implying, implied)
		}
	}

	return nil
}

func addImplications(store *storage.Storage, tx *storage.Tx, tagArgs []string) error {
	log.Infof(2, "loading settings")

	settings, err := store.Settings(tx)
	if err != nil {
		return err
	}

	tagValuePairs, warnings, err := parseTagValuePairs(tagArgs, store, tx, settings)
	if err != nil {
		return err
	}

	implyingPair := tagValuePairs[0]
	impliedPairs := tagValuePairs[1:]

	for _, impliedPair := range impliedPairs {
		log.Infof(2, "adding tag implication of '%v' to '%v'", implyingPair.TagId, impliedPair.TagId)

		if err = store.AddImplication(tx, entities.TagValuePair{implyingPair.TagId, implyingPair.ValueId}, entities.TagValuePair{impliedPair.TagId, impliedPair.ValueId}); err != nil {
			return fmt.Errorf("could not add tag implication of '%v' to '%v': %v", implyingPair, impliedPair, err)
		}
	}

	if warnings {
		return errBlank
	}

	return nil
}

func deleteImplications(store *storage.Storage, tx *storage.Tx, tagArgs []string) error {
	settings, err := store.Settings(tx)
	if err != nil {
		return err
	}

	tagValuePairs, warnings, err := parseTagValuePairs(tagArgs, store, tx, settings)
	if err != nil {
		return err
	}

	implyingPair := tagValuePairs[0]
	impliedPairs := tagValuePairs[1:]

	for _, impliedPair := range impliedPairs {
		log.Infof(2, "removing tag implication %v -> %v.", implyingPair, impliedPair)

		if err := store.DeleteImplication(tx, implyingPair, impliedPair); err != nil {
			return fmt.Errorf("could not delete tag implication of %v to %v: %v", implyingPair, impliedPair, err)
		}
	}

	if warnings {
		return errBlank
	}

	return nil
}
