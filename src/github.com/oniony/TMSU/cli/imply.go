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
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"strings"
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

// unexported

func implyExec(options Options, args []string, databasePath string) (error, warnings) {
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

	colour, err := useColour(options)
	if err != nil {
		return err, nil
	}

	if options.HasOption("--delete") {
		if len(args) < 2 {
			return fmt.Errorf("too few arguments"), nil
		}

		return deleteImplications(store, tx, args)
	}

	switch len(args) {
	case 0:
		return listImplications(store, tx, colour), nil
	case 1:
		return fmt.Errorf("tag(s) to be implied must be specified"), nil
	default:
		return addImplications(store, tx, args)
	}
}

func listImplications(store *storage.Storage, tx *storage.Tx, colour bool) error {
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
			paddingWidth := width - len(implication.ImplyingTag.Name)
			if implication.ImplyingValue.Id != 0 {
				paddingWidth -= 1 + len(implication.ImplyingValue.Name)
			}
			padding := strings.Repeat(" ", paddingWidth)

			implying := formatTagValueName(implication.ImplyingTag.Name, implication.ImplyingValue.Name, colour, false, true)
			implied := formatTagValueName(implication.ImpliedTag.Name, implication.ImpliedValue.Name, colour, true, false)

			fmt.Printf("%s%s -> %s\n", padding, implying, implied)
		}
	}

	return nil
}

func addImplications(store *storage.Storage, tx *storage.Tx, tagArgs []string) (error, warnings) {
	log.Infof(2, "loading settings")

	settings, err := store.Settings(tx)
	if err != nil {
		return err, nil
	}

	implyingTagArg := tagArgs[0]
	impliedTagArgs := tagArgs[1:]

	implyingTagName, implyingValueName := parseTagEqValueName(implyingTagArg)

	implyingTag, err := store.TagByName(tx, implyingTagName)
	if err != nil {
		return err, nil
	}
	if implyingTag == nil {
		if settings.AutoCreateTags() {
			implyingTag, err = createTag(store, tx, implyingTagName)
			if err != nil {
				return err, nil
			}
		} else {
			return NoSuchTagError{implyingTagName}, nil
		}
	}

	implyingValue, err := store.ValueByName(tx, implyingValueName)
	if err != nil {
		return err, nil
	}
	if implyingValue == nil {
		if settings.AutoCreateValues() {
			implyingValue, err = createValue(store, tx, implyingValueName)
			if err != nil {
				return err, nil
			}
		} else {
			return NoSuchValueError{implyingValueName}, nil
		}
	}

	warnings := make(warnings, 0, 10)
	for _, impliedTagArg := range impliedTagArgs {
		impliedTagName, impliedValueName := parseTagEqValueName(impliedTagArg)

		impliedTag, err := store.TagByName(tx, impliedTagName)
		if err != nil {
			return err, warnings
		}
		if impliedTag == nil {
			if settings.AutoCreateTags() {
				impliedTag, err = createTag(store, tx, impliedTagName)
				if err != nil {
					return err, warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("no such tag '%v'", impliedTagName))
				continue
			}
		}

		impliedValue, err := store.ValueByName(tx, impliedValueName)
		if err != nil {
			return err, warnings
		}
		if impliedValue == nil {
			if settings.AutoCreateValues() {
				impliedValue, err = createValue(store, tx, impliedValueName)
				if err != nil {
					return err, warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("no such value '%v'", impliedValueName))
				continue
			}
		}

		log.Infof(2, "adding tag implication of '%v' to '%v'", implyingTagArg, impliedTagArg)

		if err = store.AddImplication(tx, entities.TagIdValueIdPair{implyingTag.Id, implyingValue.Id}, entities.TagIdValueIdPair{impliedTag.Id, impliedValue.Id}); err != nil {
			return fmt.Errorf("cannot add implication of '%v' to '%v': %v", implyingTagArg, impliedTagArg, err), warnings
		}
	}

	return nil, warnings
}

func deleteImplications(store *storage.Storage, tx *storage.Tx, tagArgs []string) (error, warnings) {
	log.Infof(2, "loading settings")

	implyingTagArg := tagArgs[0]
	impliedTagArgs := tagArgs[1:]

	implyingTagName, implyingValueName := parseTagEqValueName(implyingTagArg)

	implyingTag, err := store.TagByName(tx, implyingTagName)
	if err != nil {
		return err, nil
	}
	if implyingTag == nil {
		return NoSuchTagError{implyingTagName}, nil
	}

	implyingValue, err := store.ValueByName(tx, implyingValueName)
	if err != nil {
		return err, nil
	}
	if implyingValue == nil {
		return NoSuchValueError{implyingValueName}, nil
	}

	warnings := make(warnings, 0, 10)
	for _, impliedTagArg := range impliedTagArgs {
		log.Infof(2, "removing tag implication %v -> %v.", implyingTagArg, impliedTagArg)

		impliedTagName, impliedValueName := parseTagEqValueName(impliedTagArg)

		impliedTag, err := store.TagByName(tx, impliedTagName)
		if err != nil {
			return err, warnings
		}
		if impliedTag == nil {
			warnings = append(warnings, fmt.Sprintf("no such tag '%v'", impliedTagName))
		}

		impliedValue, err := store.ValueByName(tx, impliedValueName)
		if err != nil {
			return err, warnings
		}
		if impliedValue == nil {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", impliedValueName))
		}

		if err := store.DeleteImplication(tx, entities.TagIdValueIdPair{implyingTag.Id, implyingValue.Id}, entities.TagIdValueIdPair{impliedTag.Id, impliedValue.Id}); err != nil {
			return fmt.Errorf("could not delete tag implication of %v to %v: %v", implyingTagArg, impliedTagArg, err), warnings
		}
	}

	return nil, warnings
}
