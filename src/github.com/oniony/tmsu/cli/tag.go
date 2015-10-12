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
	"bufio"
	"fmt"
	"github.com/oniony/tmsu/common/fingerprint"
	"github.com/oniony/tmsu/common/log"
	"github.com/oniony/tmsu/common/text"
	"github.com/oniony/tmsu/entities"
	"github.com/oniony/tmsu/storage"
	"io"
	"os"
	"path/filepath"
)

var TagCommand = Command{
	Name:     "tag",
	Synopsis: "Apply tags to files",
	Usages: []string{"tmsu tag [OPTION]... FILE TAG[=VALUE]...",
		`tmsu tag [OPTION]... --tags="TAG[=VALUE]..." FILE...`,
		"tmsu tag [OPTION]... --from=SOURCE FILE...",
		"tmsu tag [OPTION]... --create {TAG|=VALUE}...",
		"tmsu tag [OPTION[... -"},
	Description: `Tags the file FILE with the TAGs and VALUEs specified.

Optionally tags applied to files may be attributed with a VALUE using the TAG=VALUE syntax.

Tag and value names may consist of one or more letter, number, punctuation and symbol characters (from the corresponding Unicode categories). Tag names cannot contain the slash '/' or backslash '\' characters.

Tags will not be applied if they are already implied by tag implications. This behaviour can be overriden with the --explicit option. See the 'imply' subcommand for more information.

If a single argument of - is passed, TMSU will read lines from standard input in the format 'FILE TAG[=VALUE]...'.

Note: The equals '=' and whitespace characters must be escaped with a backslash '\' when used within a tag or value name. However, your shell may use the backslash for its own purposes: this can normally be avoided by enclosing the argument in single quotation marks or by escaping the backslash with an additional backslash '\\'.`,
	Examples: []string{"$ tmsu tag mountain1.jpg photo landscape holiday good country=france",
		"$ tmsu tag --from=mountain1.jpg mountain2.jpg",
		`$ tmsu tag --tags="landscape" field1.jpg field2.jpg`,
		"$ tmsu tag --create bad rubbish awful =2015",
		`$ tmsu tag sheep.jpg 'contains\=equals'`,
		"$ tmsu tag sheep.jpg '<tag>'"},
	Options: Options{{"--tags", "-t", "the set of tags to apply", true, ""},
		{"--recursive", "-r", "recursively apply tags to directory contents", false, ""},
		{"--from", "-f", "copy tags from the SOURCE file", true, ""},
		{"--create", "-c", "create tags without tagging any files", false, ""},
		{"--explicit", "-e", "explicitly apply tags even if they are already implied", false, ""},
		{"--force", "-F", "apply tags to non-existant or non-permissioned paths", false, ""}},
	Exec: tagExec,
}

// unexported

func tagExec(options Options, args []string, databasePath string) (error, warnings) {
	recursive := options.HasOption("--recursive")
	explicit := options.HasOption("--explicit")
	force := options.HasOption("--force")

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	switch {
	case options.HasOption("--create"):
		if len(args) == 0 {
			return fmt.Errorf("too few arguments"), nil
		}

		return createTagsValues(store, tx, args)
	case options.HasOption("--tags"):
		if len(args) < 1 {
			return fmt.Errorf("too few arguments"), nil
		}

		tagArgs := text.Tokenize(options.Get("--tags").Argument)
		if len(tagArgs) == 0 {
			return fmt.Errorf("too few arguments"), nil
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("too few arguments"), nil
		}

		return tagPaths(store, tx, tagArgs, paths, explicit, recursive, force)
	case options.HasOption("--from"):
		if len(args) < 1 {
			return fmt.Errorf("too few arguments"), nil
		}

		fromPath, err := filepath.Abs(options.Get("--from").Argument)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", fromPath, err), nil
		}

		paths := args

		return tagFrom(store, tx, fromPath, paths, explicit, recursive, force)
	case len(args) == 1 && args[0] == "-":
		return readStandardInput(store, tx, recursive, explicit, force)
	default:
		if len(args) < 2 {
			return fmt.Errorf("too few arguments"), nil
		}

		paths := args[0:1]
		tagArgs := args[1:]

		return tagPaths(store, tx, tagArgs, paths, explicit, recursive, force)
	}
}

func createTagsValues(store *storage.Storage, tx *storage.Tx, tagArgs []string) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, tagArg := range tagArgs {
		name := parseTagOrValueName(tagArg)

		if name[0] == '=' {
			name = name[1:]

			value, err := store.ValueByName(tx, name)
			if err != nil {
				return fmt.Errorf("could not check if value '%v' exists: %v", name, err), warnings
			}

			if value == nil {
				if _, err := store.AddValue(tx, name); err != nil {
					return fmt.Errorf("could not create value '%v': %v", name, err), warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("value '%v' already exists", name))
			}
		} else {
			tag, err := store.TagByName(tx, name)
			if err != nil {
				return fmt.Errorf("could not check if tag '%v' exists: %v", name, err), warnings
			}

			if tag == nil {
				if _, err := store.AddTag(tx, name); err != nil {
					return fmt.Errorf("could not create tag '%v': %v", name, err), warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("tag '%v' already exists", name))
			}
		}
	}

	return nil, warnings
}

func tagPaths(store *storage.Storage, tx *storage.Tx, tagArgs, paths []string, explicit, recursive, force bool) (error, warnings) {
	log.Infof(2, "loading settings")

	settings, err := store.Settings(tx)
	if err != nil {
		return err, nil
	}

	pairs := make(entities.TagIdValueIdPairs, 0, len(tagArgs))
	warnings := make(warnings, 0, 10)

	for _, tagArg := range tagArgs {
		tagName, valueName := parseTagEqValueName(tagArg)

		tag, err := store.TagByName(tx, tagName)
		if err != nil {
			return err, warnings
		}
		if tag == nil {
			if settings.AutoCreateTags() {
				tag, err = createTag(store, tx, tagName)
				if err != nil {
					return err, warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("no such tag '%v'", tagName))
				continue
			}
		}

		value, err := store.ValueByName(tx, valueName)
		if err != nil {
			return err, warnings
		}
		if value == nil {
			if settings.AutoCreateValues() {
				value, err = createValue(store, tx, valueName)
				if err != nil {
					return err, warnings
				}
			} else {
				warnings = append(warnings, fmt.Sprintf("no such value '%v'", valueName))
				continue
			}
		}

		pairs = append(pairs, entities.TagIdValueIdPair{tag.Id, value.Id})
	}

	for _, path := range paths {
		if err := tagPath(store, tx, path, pairs, explicit, recursive, force, settings.FileFingerprintAlgorithm(), settings.DirectoryFingerprintAlgorithm()); err != nil {
			switch {
			case os.IsPermission(err):
				warnings = append(warnings, fmt.Sprintf("%v: permisison denied", path))
			case os.IsNotExist(err):
				warnings = append(warnings, fmt.Sprintf("%v: no such file", path))
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err), warnings
			}
		}
	}

	return nil, warnings
}

func tagFrom(store *storage.Storage, tx *storage.Tx, fromPath string, paths []string, explicit, recursive, force bool) (error, warnings) {
	log.Infof(2, "loading settings")

	settings, err := store.Settings(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve settings: %v", err), nil
	}

	file, err := store.FileByPath(tx, fromPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", fromPath, err), nil
	}
	if file == nil {
		return fmt.Errorf("%v: path is not tagged", fromPath), nil
	}

	fileTags, err := store.FileTagsByFileId(tx, file.Id, true)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve filetags: %v", fromPath, err), nil
	}

	pairs := make([]entities.TagIdValueIdPair, len(fileTags))
	for index, fileTag := range fileTags {
		pairs[index] = entities.TagIdValueIdPair{fileTag.TagId, fileTag.ValueId}
	}

	warnings := make(warnings, 0, 10)

	for _, path := range paths {
		if err := tagPath(store, tx, path, pairs, explicit, recursive, force, settings.FileFingerprintAlgorithm(), settings.DirectoryFingerprintAlgorithm()); err != nil {
			switch {
			case os.IsPermission(err):
				warnings = append(warnings, fmt.Sprintf("%v: permisison denied", path))
			case os.IsNotExist(err):
				warnings = append(warnings, fmt.Sprintf("%v: no such file", path))
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err), warnings
			}
		}
	}

	return nil, warnings
}

func tagPath(store *storage.Storage, tx *storage.Tx, path string, pairs []entities.TagIdValueIdPair, explicit, recursive, force bool, fileFingerprintAlg, dirFingerprintAlg string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsNotExist(err), os.IsPermission(err):
			if !force {
				return err
			} else {
				stat = emptyStat{}
			}
		default:
			return err
		}
	}

	log.Infof(2, "%v: checking if file exists in database", path)

	file, err := store.FileByPath(tx, absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		log.Infof(2, "%v: creating fingerprint", path)

		fingerprint, err := fingerprint.Create(path, fileFingerprintAlg, dirFingerprintAlg)
		if err != nil {
			if !force || !(os.IsNotExist(err) || os.IsPermission(err)) {
				return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
			}
		}

		log.Infof(2, "%v: adding file.", path)

		file, err = store.AddFile(tx, path, fingerprint, stat.ModTime(), int64(stat.Size()), stat.IsDir())
		if err != nil {
			return fmt.Errorf("%v: could not add file to database: %v", path, err)
		}
	}

	if !explicit {
		pairs, err = removeAlreadyAppliedTagValuePairs(store, tx, pairs, file)
		if err != nil {
			return fmt.Errorf("%v: could not remove applied tags: %v", path, err)
		}
	}

	log.Infof(2, "%v: applying tags.", path)

	for _, pair := range pairs {
		if _, err = store.AddFileTag(tx, file.Id, pair.TagId, pair.ValueId); err != nil {
			return fmt.Errorf("%v: could not apply tags: %v", file.Path(), err)
		}
	}

	if recursive && stat.IsDir() {
		if err = tagRecursively(store, tx, path, pairs, explicit, force, fileFingerprintAlg, dirFingerprintAlg); err != nil {
			return err
		}
	}

	return nil
}

func readStandardInput(store *storage.Storage, tx *storage.Tx, recursive, explicit, force bool) (error, warnings) {
	reader := bufio.NewReader(os.Stdin)

	warnings := make(warnings, 0, 10)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return err, warnings
		}

		words := text.Tokenize(line[0 : len(line)-1])

		if len(words) < 2 {
			warnings = append(warnings, fmt.Sprintf("too few arguments"))
			continue
		}

		path := words[0]
		tagArgs := words[1:]

		err, commandWarnings := tagPaths(store, tx, tagArgs, []string{path}, explicit, recursive, force)
		if err != nil {
			warnings = append(warnings, err.Error())
		}
		if commandWarnings != nil {
			warnings = append(warnings, commandWarnings...)
		}
	}

	return nil, warnings
}

func tagRecursively(store *storage.Storage, tx *storage.Tx, path string, pairs []entities.TagIdValueIdPair, explicit, force bool, fileFingerprintAlg, dirFingerprintAlg string) error {
	osFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%v: could not open path: %v", path, err)
	}

	childNames, err := osFile.Readdirnames(0)
	osFile.Close()
	if err != nil {
		return fmt.Errorf("%v: could not retrieve directory contents: %v", path, err)
	}

	for _, childName := range childNames {
		childPath := filepath.Join(path, childName)

		if err = tagPath(store, tx, childPath, pairs, explicit, true, force, fileFingerprintAlg, dirFingerprintAlg); err != nil {
			return err
		}
	}

	return nil
}

func removeAlreadyAppliedTagValuePairs(store *storage.Storage, tx *storage.Tx, pairs []entities.TagIdValueIdPair, file *entities.File) ([]entities.TagIdValueIdPair, error) {
	log.Infof(2, "%v: determining existing file-tags", file.Path())

	existingFileTags, err := store.FileTagsByFileId(tx, file.Id, false)
	if err != nil {
		return nil, fmt.Errorf("%v: could not determine file's tags: %v", file.Path(), err)
	}

	log.Infof(2, "%v: determining implied tags", file.Path())

	newImplications, err := store.ImplicationsFor(tx, pairs...)
	if err != nil {
		return nil, fmt.Errorf("%v: could not determine implied tags: %v", file.Path(), err)
	}

	log.Infof(2, "%v: revising set of tags to apply", file.Path())

	revisedPairs := make([]entities.TagIdValueIdPair, 0, len(pairs))
	for _, pair := range pairs {
		predicate := func(ft entities.FileTag) bool {
			return ft.TagId == pair.TagId && ft.ValueId == pair.ValueId
		}

		if existingFileTags.Any(predicate) {
			continue
		}

		if newImplications.Implies(pair) {
			continue
		}

		revisedPairs = append(revisedPairs, pair)
	}

	return revisedPairs, nil
}
