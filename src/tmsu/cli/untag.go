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
	"path/filepath"
	"strings"
	"tmsu/common/log"
	"tmsu/entities"
	"tmsu/storage"
)

var UntagCommand = Command{
	Name:     "untag",
	Synopsis: "Remove tags from files",
	Description: `tmsu untag [OPTION]... FILE TAG...
tmsu untag [OPTION]... --all FILE...
tmsu untag [OPTION]... --tags="TAG..." FILE...

Disassociates FILE with the TAGs specified.

Examples:

    $ tmsu untag mountain.jpg hill
    $ tmsu untag --all mountain-copy.jpg
    $ tmsu untag --tags="river underwater" forest.jpg desert.jpg`,
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

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if options.HasOption("--all") {
		if len(args) < 1 {
			return fmt.Errorf("files to untag must be specified.")
		}

		paths := args

		if err := untagPathsAll(store, paths, recursive); err != nil {
			return err
		}
	} else if options.HasOption("--tags") {
		tagNames := strings.Fields(options.Get("--tags").Argument)
		if len(tagNames) == 0 {
			return fmt.Errorf("set of tags to apply must be specified")
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("at least one file to untag must be specified")
		}

		tagIds, err := lookupTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err := untagPaths(store, paths, tagIds, recursive); err != nil {
			return err
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("tag to remove and files to untag must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		tagIds, err := lookupTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err := untagPath(store, path, tagIds, recursive); err != nil {
			return err
		}
	}

	return nil
}

func lookupTagIds(store *storage.Storage, names []string) ([]uint, error) {
	tags, err := store.TagsByNames(names)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve tags: %v", err)
	}

	for _, name := range names {
		if !tags.Any(func(tag *entities.Tag) bool { return tag.Name == name }) {
			return nil, fmt.Errorf("no such tag '%v'", name)
		}
	}

	tagIds := make([]uint, len(tags))
	for index, tag := range tags {
		tagIds[index] = tag.Id
	}

	return tagIds, nil
}

func untagPathsAll(store *storage.Storage, paths []string, recursive bool) error {
	for _, path := range paths {
		if err := untagPathAll(store, path, recursive); err != nil {
			return err
		}
	}

	return nil
}

func untagPaths(store *storage.Storage, paths []string, tagIds []uint, recursive bool) error {
	for _, path := range paths {
		if err := untagPath(store, path, tagIds, recursive); err != nil {
			return err
		}
	}

	return nil
}

func untagPathAll(store *storage.Storage, path string, recursive bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		return fmt.Errorf("%v: file is not tagged.", path)
	}

	log.Infof(2, "%v: removing all tags.", file.Path())

	if err := store.RemoveFileTagsByFileId(file.Id); err != nil {
		return fmt.Errorf("%v: could not remove file's tags: %v", file.Path(), err)
	}

	if err := removeUntaggedFile(store, file); err != nil {
		return err
	}

	if recursive {
		childFiles, err := store.FilesByDirectory(file.Path())
		if err != nil {
			return fmt.Errorf("%v: could not retrieve files for directory: %v", file.Path())
		}

		for _, childFile := range childFiles {
			if err := store.RemoveFileTagsByFileId(childFile.Id); err != nil {
				return fmt.Errorf("%v: could not remove file's tags: %v", childFile.Path(), err)
			}

			if err := removeUntaggedFile(store, childFile); err != nil {
				return err
			}

		}
	}

	return nil
}

func untagPath(store *storage.Storage, path string, tagIds []uint, recursive bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		return fmt.Errorf("%v: file is not tagged.", path)
	}

	for _, tagId := range tagIds {
		log.Infof(2, "%v: unapplying tag #%v.", file.Path(), tagId)

		//TODO value id
		if err := store.RemoveFileTag(file.Id, tagId, 0); err != nil {
			return fmt.Errorf("%v: could not remove tag #%v: %v", file.Path(), tagId, err)
		}
	}

	if err := removeUntaggedFile(store, file); err != nil {
		return err
	}

	if recursive {
		childFiles, err := store.FilesByDirectory(file.Path())
		if err != nil {
			return fmt.Errorf("%v: could not retrieve files for directory: %v", file.Path())
		}

		for _, childFile := range childFiles {
			for _, tagId := range tagIds {
				log.Infof(2, "%v: unapplying tag #%v.", childFile.Path(), tagId)

				//TODO value id
				if err := store.RemoveFileTag(childFile.Id, tagId, 0); err != nil {
					return fmt.Errorf("%v: could not remove tag #%v: %v", childFile.Path(), tagId, err)
				}
			}

			if err := removeUntaggedFile(store, childFile); err != nil {
				return err
			}
		}
	}

	return nil
}

func removeUntaggedFile(store *storage.Storage, file *entities.File) error {
	log.Infof(2, "%v: identifying whether file is tagged.", file.Path())

	filetagCount, err := store.FileTagCountByFileId(file.Id)
	if err != nil {
		return fmt.Errorf("%v: could not get tag count: %v", file.Path(), err)
	}

	if filetagCount != 0 {
		return nil
	}

	log.Infof(2, "%v: removing untagged file.", file.Path())

	err = store.RemoveFile(file.Id)
	if err != nil {
		return fmt.Errorf("%v: could not remove file: %v", file.Path(), err)
	}

	return nil
}
