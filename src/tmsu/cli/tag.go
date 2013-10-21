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
	"os"
	"path/filepath"
	"strings"
	"time"
	"tmsu/entities"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
)

var TagCommand = Command{
	Name:     "tag",
	Synopsis: "Apply tags to files",
	Description: `tmsu tag [OPTION]... FILE TAG...
tmsu tag [OPTION]... --tags "TAG..." FILE...
tmsu tag [OPTION]... --from FILE FILE...

Tags the file FILE with the tag(s) specified.`,
	Options: Options{{"--tags", "-t", "the set of tags to apply", true, ""},
		{"--recursive", "-r", "recursively apply tags to directory contents", false, ""},
		{"--from", "-f", "copy tags from the specified file", true, ""}},
	Exec: tagExec,
}

func tagExec(options Options, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments.")
	}

	recursive := options.HasOption("--recursive")

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	switch {
	case options.HasOption("--tags"):
		tagNames := strings.Fields(options.Get("--tags").Argument)
		if len(tagNames) == 0 {
			return fmt.Errorf("set of tags to apply must be specified")
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("at least one file to tag must be specified")
		}

		tagIds, err := lookupOrCreateTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err := tagPaths(store, paths, tagIds, recursive); err != nil {
			return err
		}
	case options.HasOption("--from"):
		fromPath, err := filepath.Abs(options.Get("--from").Argument)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", fromPath, err)
		}

		tags, err := store.TagsForPath(fromPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve tags: %v", fromPath)
		}

		tagIds := make([]uint, len(tags))
		for index, tag := range tags {
			tagIds[index] = tag.Id
		}

		for _, path := range args {
			if err = tagPath(store, path, tagIds, recursive); err != nil {
				return err
			}
		}
	default:
		if len(args) < 2 {
			return fmt.Errorf("file to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		err := ValidateTagNames(tagNames)
		if err != nil {
			return err
		}

		tagIds, err := lookupOrCreateTagIds(store, tagNames)
		if err != nil {
			return err
		}

		if err = tagPath(store, path, tagIds, recursive); err != nil {
			return err
		}
	}

	return nil
}

func lookupOrCreateTagIds(store *storage.Storage, names []string) ([]uint, error) {
	tags, err := store.TagsByNames(names)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve tags %v: %v", names, err)
	}

	for _, name := range names {
		if !tags.Any(func(tag *entities.Tag) bool { return tag.Name == name }) {
			log.Warnf("New tag '%v'.", name)

			tag, err := store.AddTag(name)
			if err != nil {
				return nil, fmt.Errorf("could not add tag '%v': %v", name, err)
			}

			tags = append(tags, tag)
		}
	}

	log.Suppf("retrieving tag implications")

	tagIds := make([]uint, len(tags))
	for index, tag := range tags {
		tagIds[index] = tag.Id
	}

	implications, err := store.ImplicationsForTags(tagIds...)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve implied tags: %v", err)
	}

	for _, implication := range implications {
		if !contains(tagIds, implication.ImpliedTag.Id) {
			log.Infof("tag '%v' is implied.", implication.ImpliedTag.Name)
			tagIds = append(tagIds, implication.ImpliedTag.Id)
		}
	}

	return tagIds, nil
}

func tagPaths(store *storage.Storage, paths []string, tagIds []uint, recursive bool) error {
	for _, path := range paths {
		if err := tagPath(store, path, tagIds, recursive); err != nil {
			return err
		}
	}

	return nil
}

func tagPath(store *storage.Storage, path string, tagIds []uint, recursive bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", path, err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return fmt.Errorf("%v: permisison denied", path)
		case os.IsNotExist(err):
			return fmt.Errorf("%v: no such file", path)
		default:
			return fmt.Errorf("%v: could not stat file: %v", path, err)
		}
	}

	file, err := store.FileByPath(absPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", path, err)
	}
	if file == nil {
		file, err = addFile(store, absPath, stat.ModTime(), uint(stat.Size()), stat.IsDir())
		if err != nil {
			return fmt.Errorf("%v: could not add file: %v", path, err)
		}
	}

	log.Suppf("%v: applying tags.", file.Path())

	if err = store.AddFileTags(file.Id, tagIds); err != nil {
		return fmt.Errorf("%v: could not apply tags: %v", file.Path(), err)
	}

	if recursive && stat.IsDir() {
		if err = tagRecursively(store, path, tagIds); err != nil {
			return err
		}
	}

	return nil
}

func tagRecursively(store *storage.Storage, path string, tagIds []uint) error {
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

		if err = tagPath(store, childPath, tagIds, true); err != nil {
			return err
		}
	}

	return nil
}

func addFile(store *storage.Storage, path string, modTime time.Time, size uint, isDir bool) (*entities.File, error) {
	log.Suppf("%v: adding file.", path)

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, fmt.Errorf("%v: could not create fingerprint: %v", path, err)
	}

	file, err := store.AddFile(path, fingerprint, modTime, int64(size), isDir)
	if err != nil {
		return nil, fmt.Errorf("%v: could not add file to database: %v", path, err)
	}

	return file, nil
}

func contains(tagIds []uint, tagId uint) bool {
	for _, id := range tagIds {
		if id == tagId {
			return true
		}
	}

	return false
}
