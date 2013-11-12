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
tmsu tag [OPTION]... --tags="TAG..." FILE...
tmsu tag [OPTION]... --from=FILE FILE...
tmsu tag [OPTION]... --create="TAG..."

Tags the file FILE with the tag(s) specified.`,
	Options: Options{{"--tags", "-t", "the set of tags to apply", true, ""},
		{"--recursive", "-r", "recursively apply tags to directory contents", false, ""},
		{"--from", "-f", "copy tags from the specified file", true, ""},
		{"--create", "-c", "create a tag without tagging any files", true, ""}},
	Exec: tagExec,
}

func tagExec(options Options, args []string) error {
	recursive := options.HasOption("--recursive")

	switch {
	case options.HasOption("--create"):
		tagNames := strings.Fields(options.Get("--create").Argument)
		if len(tagNames) == 0 {
			return fmt.Errorf("set of tags to create must be specified")
		}

		if len(args) > 0 {
			return fmt.Errorf("too many arguments")
		}

		if err := createTags(tagNames); err != nil {
			return err
		}
	case options.HasOption("--tags"):
		if len(args) < 1 {
			return fmt.Errorf("files to tag must be specified")
		}

		tagNames := strings.Fields(options.Get("--tags").Argument)
		if len(tagNames) == 0 {
			return fmt.Errorf("set of tags to apply must be specified")
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("at least one file to tag must be specified")
		}

		if err := tagPaths(tagNames, paths, recursive); err != nil {
			return err
		}
	case options.HasOption("--from"):
		if len(args) < 1 {
			return fmt.Errorf("files to tag must be specified")
		}

		fromPath, err := filepath.Abs(options.Get("--from").Argument)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", fromPath, err)
		}

		paths := args

		if err := tagFrom(fromPath, paths, recursive); err != nil {
			return err
		}
	default:
		if len(args) < 2 {
			return fmt.Errorf("file to tag and tags to apply must be specified.")
		}

		paths := args[0:1]
		tagNames := args[1:]

		if err := tagPaths(tagNames, paths, recursive); err != nil {
			return err
		}
	}

	return nil
}

func createTags(names []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	tags, err := store.TagsByNames(names)
	if err != nil {
		return fmt.Errorf("could not retrieve tags %v: %v", names, err)
	}

	if len(tags) > 0 {
		return fmt.Errorf("tags already exists: %v", tags[0].Name)
	}

	for _, name := range names {
		_, err := store.AddTag(name)
		if err != nil {
			return fmt.Errorf("could not add tag '%v': %v", name, err)
		}
	}

	return nil
}

func tagPaths(tagNames, paths []string, recursive bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	tagIds, err := lookupOrCreateTagIds(store, tagNames)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := tagPath(store, path, tagIds, recursive); err != nil {
			return err
		}
	}

	return nil
}

func tagFrom(fromPath string, paths []string, recursive bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	tags, err := store.TagsForPath(fromPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve tags: %v", fromPath)
	}

	tagIds := make([]uint, len(tags))
	for index, tag := range tags {
		tagIds[index] = tag.Id
	}

	for _, path := range paths {
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

	log.Infof(2, "retrieving tag implications")

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
			log.Warnf("tag '%v' is implied.", implication.ImpliedTag.Name)
			tagIds = append(tagIds, implication.ImpliedTag.Id)
		}
	}

	return tagIds, nil
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

	log.Infof(2, "%v: checking if file exists", absPath)

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

	log.Infof(2, "%v: applying tags.", file.Path())

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
	log.Infof(2, "%v: creating fingerprint", path)

	fingerprint, err := fingerprint.Create(path)
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

func contains(tagIds []uint, tagId uint) bool {
	for _, id := range tagIds {
		if id == tagId {
			return true
		}
	}

	return false
}
