/*
Copyright 2011-2014 Paul Ruane.

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
	"path/filepath"
	"strings"
	"tmsu/common/log"
	"tmsu/storage"
)

var UntagCommand = Command{
	Name:     "untag",
	Synopsis: "Remove tags from files",
	Description: `tmsu untag [OPTION]... FILE TAG[=VALUE]...
tmsu untag [OPTION]... --all FILE...
tmsu untag [OPTION]... --tags="TAG[=VALUE]..." FILE...

Disassociates FILE with the TAGs specified.

Examples:

    $ tmsu untag mountain.jpg hill county=germany
    $ tmsu untag --all mountain-copy.jpg
    $ tmsu untag --tags="river underwater year=2014" forest.jpg desert.jpg`,
	Options: Options{{"--all", "-a", "strip each file of all tags", false, ""},
		{"--tags", "-t", "the set of tags to remove", true, ""},
		{"--recursive", "-r", "recursively remove tags from directory contents", false, ""}},
	Exec: untagExec,
}

func untagExec(options Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no arguments specified.")
	}

	recursive := options.HasOption("--recursive")

	if options.HasOption("--all") {
		if len(args) < 1 {
			return fmt.Errorf("files to untag must be specified.")
		}

		paths := args

		if err := untagPathsAll(paths, recursive); err != nil {
			return err
		}
	} else if options.HasOption("--tags") {
		tagArgs := strings.Fields(options.Get("--tags").Argument)
		if len(tagArgs) == 0 {
			return fmt.Errorf("set of tags to apply must be specified")
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("at least one file to untag must be specified")
		}

		if err := untagPaths(paths, tagArgs, recursive); err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("tags to remove and files to untag must be specified.")
		}

		paths := args[0:1]
		tagArgs := args[1:]

		if err := untagPaths(paths, tagArgs, recursive); err != nil {
			return err
		}
	}

	return nil
}

func untagPathsAll(paths []string, recursive bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()
	defer store.Commit()

	wereErrors := false
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file == nil {
			log.Warnf("%v: file is not tagged.", path)
			wereErrors = true
			continue
		}

		log.Infof(2, "%v: removing all tags.", file.Path())

		if err := store.DeleteFileTagsByFileId(file.Id); err != nil {
			return fmt.Errorf("%v: could not remove file's tags: %v", file.Path(), err)
		}

		if recursive {
			childFiles, err := store.FilesByDirectory(file.Path())
			if err != nil {
				return fmt.Errorf("%v: could not retrieve files for directory: %v", file.Path())
			}

			for _, childFile := range childFiles {
				if err := store.DeleteFileTagsByFileId(childFile.Id); err != nil {
					return fmt.Errorf("%v: could not remove file's tags: %v", childFile.Path(), err)
				}
			}
		}
	}

	if wereErrors {
		return blankError
	}

	return nil
}

func untagPaths(paths, tagArgs []string, recursive bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()
	defer store.Commit()

	tagValuePairs := make([]TagValuePair, 0, 10)
	wereErrors := false
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
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
		}
		if tag == nil {
			log.Warnf("no such tag '%v'", tagName)
			wereErrors = true
		}

		value, err := store.ValueByName(valueName)
		if err != nil {
			return fmt.Errorf("could not retrieve value '%v': %v", valueName, err)
		}
		if value == nil {
			log.Warnf("no such value '%v'", valueName)
			wereErrors = true
		}

		if tag != nil && value != nil {
			tagValuePairs = append(tagValuePairs, TagValuePair{tag.Id, value.Id})
		}
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file == nil {
			log.Warnf("%v: file is not tagged", path)
			wereErrors = true
			continue
		}

		log.Infof(2, "%v: unapplying tags.", file.Path())

		for _, tagValuePair := range tagValuePairs {
			if err := store.DeleteFileTag(file.Id, tagValuePair.TagId, tagValuePair.ValueId); err != nil {
				return fmt.Errorf("%v: could not remove tag #%v, value #%v: %v", file.Path(), tagValuePair.TagId, tagValuePair.ValueId, err)
			}
		}

		if recursive {
			childFiles, err := store.FilesByDirectory(file.Path())
			if err != nil {
				return fmt.Errorf("%v: could not retrieve files for directory: %v", file.Path())
			}

			for _, childFile := range childFiles {
				log.Infof(2, "%v: unapplying tags.", childFile.Path())

				for _, tagValuePair := range tagValuePairs {
					if err := store.DeleteFileTag(childFile.Id, tagValuePair.TagId, tagValuePair.ValueId); err != nil {
						return fmt.Errorf("%v: could not remove tag #%v, value #%v: %v", childFile.Path(), tagValuePair.TagId, tagValuePair.ValueId, err)
					}
				}
			}
		}
	}

	if wereErrors {
		return blankError
	}

	return nil
}
