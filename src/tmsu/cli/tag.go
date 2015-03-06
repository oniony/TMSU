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
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tmsu/common/fingerprint"
	"tmsu/common/log"
	"tmsu/common/text"
	"tmsu/entities"
	"tmsu/storage"
)

var TagCommand = Command{
	Name:     "tag",
	Synopsis: "Apply tags to files",
	Usages: []string{"tmsu tag [OPTION]... FILE TAG[=VALUE]...",
		`tmsu tag [OPTION]... --tags="TAG[=VALUE]..." FILE...`,
		"tmsu tag [OPTION]... --from=SOURCE FILE...",
		"tmsu tag [OPTION]... --create TAG[=VALUE]...",
		"tmsu tag [OPTION[... -"},
	Description: `Tags the file FILE with the TAGs specified. If no TAG is specified then all tags are listed.

Tag names may consist of one or more letter, number, punctuation and symbol characters (from the corresponding Unicode categories). Tag names may not contain whitespace characters, the comparison operator symbols ('=', '<' and '>"), parentheses ('(' and ')'), commas (',') or the slash symbol ('/'). In addition, the tag names '.' and '..' are not valid.

Optionally tags applied to files may be attributed with a VALUE using the TAG=VALUE syntax.

If a single argument of - is passed, TMSU will read lines from standard input in the format 'FILE TAG[=VALUE]...'.`,
	Examples: []string{"$ tmsu tag mountain1.jpg photo landscape holiday good country=france",
		"$ tmsu tag --from=mountain1.jpg mountain2.jpg",
		`$ tmsu tag --tags="landscape" field1.jpg field2.jpg`,
		"$ tmsu tag --create bad rubbish awful"},
	Options: Options{{"--tags", "-t", "the set of tags to apply", true, ""},
		{"--recursive", "-r", "recursively apply tags to directory contents", false, ""},
		{"--from", "-f", "copy tags from the SOURCE file", true, ""},
		{"--create", "-c", "create tags without tagging any files", false, ""},
		{"--explicit", "-e", "explicitly apply tags even if they are already implied", false, ""},
		{"--force", "-F", "create tags for non-existant or permissioned paths", false, ""}},
	Exec: tagExec,
}

func tagExec(store *storage.Storage, options Options, args []string) error {
	recursive := options.HasOption("--recursive")
	explicit := options.HasOption("--explicit")
	force := options.HasOption("--force")

	if err := store.Begin(); err != nil {
		return err
	}
	defer store.Commit()

	switch {
	case options.HasOption("--create"):
		if len(args) == 0 {
			return fmt.Errorf("too few arguments")
		}

		if err := createTags(store, args); err != nil {
			return err
		}
	case options.HasOption("--tags"):
		if len(args) < 1 {
			return fmt.Errorf("too few arguments")
		}

		tagArgs := strings.Fields(options.Get("--tags").Argument)
		if len(tagArgs) == 0 {
			return fmt.Errorf("too few arguments")
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("too few arguments")
		}

		if err := tagPaths(store, tagArgs, paths, explicit, recursive, force); err != nil {
			return err
		}
	case options.HasOption("--from"):
		if len(args) < 1 {
			return fmt.Errorf("too few arguments")
		}

		fromPath, err := filepath.Abs(options.Get("--from").Argument)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", fromPath, err)
		}

		paths := args

		if err := tagFrom(store, fromPath, paths, explicit, recursive, force); err != nil {
			return err
		}
	case len(args) == 1 && args[0] == "-":
		if err := readStandardInput(store, recursive, explicit, force); err != nil {
			return err
		}
	default:
		if len(args) < 2 {
			return fmt.Errorf("too few arguments")
		}

		paths := args[0:1]
		tagArgs := args[1:]

		if err := tagPaths(store, tagArgs, paths, explicit, recursive, force); err != nil {
			return err
		}
	}

	return nil
}

func createTags(store *storage.Storage, tagNames []string) error {
	wereErrors := false
	for _, tagName := range tagNames {
		tag, err := store.TagByName(tagName)
		if err != nil {
			return fmt.Errorf("could not check if tag '%v' exists: %v", tagName, err)
		}

		if tag == nil {
			log.Infof(2, "adding tag '%v'.", tagName)

			_, err := store.AddTag(tagName)
			if err != nil {
				return fmt.Errorf("could not add tag '%v': %v", tagName, err)
			}
		} else {
			log.Warnf("tag '%v' already exists", tagName)
			wereErrors = true
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func tagPaths(store *storage.Storage, tagArgs, paths []string, explicit, recursive, force bool) error {
	log.Infof(2, "loading settings")

	settings, err := store.Settings()
	if err != nil {
		return err
	}

	wereErrors := false
	tagValuePairs := make([]tagValuePair, 0, 10)
	for _, tagArg := range tagArgs {
		var tagName, valueName string
		index := strings.Index(tagArg, "=")

		switch index {
		case -1, 0:
			tagName = tagArg
		default:
			tagName = tagArg[0:index]
			valueName = tagArg[index+1 : len(tagArg)]
		}

		tag, err := store.TagByName(tagName)
		if err != nil {
			return err
		}
		if tag == nil {
			if settings.AutoCreateTags() {
				tag, err = createTag(store, tagName)
				if err != nil {
					return err
				}
			} else {
				log.Warnf("no such tag '%v'.", tagName)
				wereErrors = true
				continue
			}
		}

		value, err := store.ValueByName(valueName)
		if err != nil {
			return err
		}
		if value == nil {
			if settings.AutoCreateValues() {
				value, err = createValue(store, valueName)
				if err != nil {
					return err
				}
			} else {
				log.Warnf("no such value '%v'.", valueName)
				wereErrors = true
				continue
			}
		}

		tagValuePairs = append(tagValuePairs, tagValuePair{tag.Id, value.Id})
	}

	for _, path := range paths {
		if err := tagPath(store, path, tagValuePairs, explicit, recursive, force, settings.FileFingerprintAlgorithm(), settings.DirectoryFingerprintAlgorithm()); err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("%v: permisison denied", path)
				wereErrors = true
			case os.IsNotExist(err):
				log.Warnf("%v: no such file", path)
				wereErrors = true
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err)
			}
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func tagFrom(store *storage.Storage, fromPath string, paths []string, explicit, recursive, force bool) error {
	log.Infof(2, "loading settings")

	settings, err := store.Settings()
	if err != nil {
		return fmt.Errorf("could not retrieve settings: %v", err)
	}

	file, err := store.FileByPath(fromPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", fromPath, err)
	}
	if file == nil {
		return fmt.Errorf("%v: path is not tagged")
	}

	fileTags, err := store.FileTagsByFileId(file.Id, true)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve filetags: %v", fromPath, err)
	}

	tagValuePairs := make([]tagValuePair, len(fileTags))
	for index, fileTag := range fileTags {
		tagValuePairs[index] = tagValuePair{fileTag.TagId, fileTag.ValueId}
	}

	wereErrors := false
	for _, path := range paths {
		if err := tagPath(store, path, tagValuePairs, explicit, recursive, force, settings.FileFingerprintAlgorithm(), settings.DirectoryFingerprintAlgorithm()); err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("%v: permisison denied", path)
				wereErrors = true
			case os.IsNotExist(err):
				log.Warnf("%v: no such file", path)
				wereErrors = true
			default:
				return fmt.Errorf("%v: could not stat file: %v", path, err)
			}
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func tagPath(store *storage.Storage, path string, tagValuePairs []tagValuePair, explicit, recursive, force bool, fileFingerprintAlg, dirFingerprintAlg string) error {
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

	log.Infof(2, "%v: checking if file exists", path)

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		file, err = addFile(store, absPath, stat.ModTime(), uint(stat.Size()), stat.IsDir(), fileFingerprintAlg, dirFingerprintAlg)
		if err != nil {
			return fmt.Errorf("%v: could not add file: %v", path, err)
		}
	}

	if !explicit {
		tagValuePairs, err = removeAlreadyAppliedTagValuePairs(store, tagValuePairs, file)
		if err != nil {
			return fmt.Errorf("%v: could not remove applied tags: %v", path, err)
		}
	}

	log.Infof(2, "%v: applying tags.", path)

	for _, tagValuePair := range tagValuePairs {
		if _, err = store.AddFileTag(file.Id, tagValuePair.TagId, tagValuePair.ValueId); err != nil {
			return fmt.Errorf("%v: could not apply tags: %v", file.Path(), err)
		}
	}

	if recursive && stat.IsDir() {
		if err = tagRecursively(store, path, tagValuePairs, explicit, force, fileFingerprintAlg, dirFingerprintAlg); err != nil {
			return err
		}
	}

	return nil
}

func readStandardInput(store *storage.Storage, recursive, explicit, force bool) error {
	reader := bufio.NewReader(os.Stdin)

	wereErrors := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		words := text.Tokenize(line[0 : len(line)-1])

		if len(words) < 2 {
			log.Warnf("too few arguments")
			wereErrors = true
			continue
		}

		path := words[0]
		tagArgs := words[1:]

		if err := tagPaths(store, tagArgs, []string{path}, explicit, recursive, force); err != nil {
			log.Warnf("%v: %v", path, err)
			wereErrors = true
		}
	}

	if wereErrors {
		return errBlank
	}

	return nil
}

func tagRecursively(store *storage.Storage, path string, tagValuePairs []tagValuePair, explicit, force bool, fileFingerprintAlg, dirFingerprintAlg string) error {
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

		if err = tagPath(store, childPath, tagValuePairs, explicit, true, force, fileFingerprintAlg, dirFingerprintAlg); err != nil {
			return err
		}
	}

	return nil
}

func addFile(store *storage.Storage, path string, modTime time.Time, size uint, isDir bool, fileFingerprintAlg, dirFingerprintAlg string) (*entities.File, error) {
	log.Infof(2, "%v: creating fingerprint", path)

	fingerprint, err := fingerprint.Create(path, fileFingerprintAlg, dirFingerprintAlg)
	if err != nil {
		return nil, fmt.Errorf("%v: could not create fingerprint: %v", path, err)
	}

	log.Infof(2, "%v: adding file.", path)

	file, err := store.AddFile(path, fingerprint, modTime, int64(size), isDir)
	if err != nil {
		return nil, fmt.Errorf("%v: could not add file to database: %v", path, err)
	}

	return file, nil
}

func removeAlreadyAppliedTagValuePairs(store *storage.Storage, tagValuePairs []tagValuePair, file *entities.File) ([]tagValuePair, error) {
	log.Infof(2, "%v: determining existing file-tags", file.Path())

	existingFileTags, err := store.FileTagsByFileId(file.Id, false)
	if err != nil {
		return nil, fmt.Errorf("%v: could not determine file's tags: %v", file.Path(), err)
	}

	log.Infof(2, "%v: determining implied tags", file.Path())

	tagIds := make(entities.TagIds, len(tagValuePairs))
	for index, tagValuePair := range tagValuePairs {
		tagIds[index] = tagValuePair.TagId
	}

	newlyImpliedTags, err := store.ImplicationsForTags(tagIds...)
	if err != nil {
		return nil, fmt.Errorf("%v: could not determine implied tags: %v", file.Path(), err)
	}

	log.Infof(2, "%v: revising set of tags to apply", file.Path())

	revisedTagValuePairs := make([]tagValuePair, 0, len(tagValuePairs))
	for _, tagValuePair := range tagValuePairs {
		if existingFileTags.Contains(tagValuePair.TagId, tagValuePair.ValueId) {
			continue
		}

		if tagValuePair.ValueId == 0 && newlyImpliedTags.Implies(tagValuePair.TagId) {
			continue
		}

		revisedTagValuePairs = append(revisedTagValuePairs, tagValuePair)
	}

	return revisedTagValuePairs, nil
}
