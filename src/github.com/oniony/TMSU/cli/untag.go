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
	_path "github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/common/text"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
)

var UntagCommand = Command{
	Name:     "untag",
	Synopsis: "Remove tags from files",
	Usages: []string{"tmsu untag [OPTION]... FILE TAG[=VALUE]...",
		"tmsu untag [OPTION]... --all FILE...",
		`tmsu untag [OPTION]... --tags="TAG[=VALUE]..." FILE...`},
	Description: "Disassociates FILE with the TAGs specified.",
	Examples: []string{"$ tmsu untag mountain.jpg hill county=germany",
		"$ tmsu untag --all mountain-copy.jpg",
		`$ tmsu untag --tags="river underwater year=2017" forest.jpg desert.jpg`},
	Options: Options{{"--all", "-a", "strip each file of all tags", false, ""},
		{"--tags", "-t", "the set of tags to remove", true, ""},
		{"--recursive", "-r", "recursively remove tags from directory contents", false, ""},
		{"--no-dereference", "-P", "do not follow symbolic links (untag the link itself)", false, ""}},
	Exec: untagExec,
}

// unexported

func untagExec(options Options, args []string, databasePath string) (error, warnings) {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments"), nil
	}

	recursive := options.HasOption("--recursive")
	followSymlinks := !options.HasOption("--no-dereference")

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

	if options.HasOption("--all") {
		if len(args) < 1 {
			return fmt.Errorf("files to untag must be specified"), nil
		}

		paths := args

		return untagPathsAll(store, tx, paths, recursive, followSymlinks)
	} else if options.HasOption("--tags") {
		tagArgs := text.Tokenize(options.Get("--tags").Argument)
		if len(tagArgs) == 0 {
			return fmt.Errorf("set of tags to apply must be specified"), nil
		}

		paths := args
		if len(paths) < 1 {
			return fmt.Errorf("at least one file to untag must be specified"), nil
		}

		return untagPaths(store, tx, paths, tagArgs, recursive, followSymlinks)
	} else {
		if len(args) < 2 {
			return fmt.Errorf("tags to remove and files to untag must be specified"), nil
		}

		paths := args[0:1]
		tagArgs := args[1:]

		return untagPaths(store, tx, paths, tagArgs, recursive, followSymlinks)
	}
}

func untagPathsAll(store *storage.Storage, tx *storage.Tx, paths []string, recursive, followSymlinks bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err), warnings
		}

		log.Infof(2, "%v: resolving path", path)

		stat, err := os.Lstat(absPath)
		if err != nil {
			switch {
			case os.IsNotExist(err), os.IsPermission(err):
				// ignore
			default:
				return err, nil
			}
		} else if stat.Mode()&os.ModeSymlink != 0 && followSymlinks {
			absPath, err = _path.Dereference(absPath)
			if err != nil {
				return err, nil
			}
		}

		file, err := store.FileByPath(tx, absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err), warnings
		}
		if file == nil {
			warnings = append(warnings, fmt.Sprintf("%v: file is not tagged.", path))
			continue
		}

		log.Infof(2, "%v: removing all tags.", path)

		if err := store.DeleteFileTagsByFileId(tx, file.Id); err != nil {
			return fmt.Errorf("%v: could not remove file's tags: %v", path, err), warnings
		}

		if recursive {
			childFiles, err := store.FilesByDirectory(tx, file.Path())
			if err != nil {
				return fmt.Errorf("%v: could not retrieve files for directory: %v", path, err), warnings
			}

			for _, childFile := range childFiles {
				if err := store.DeleteFileTagsByFileId(tx, childFile.Id); err != nil {
					return fmt.Errorf("%v: could not remove file's tags: %v", childFile.Path(), err), warnings
				}
			}
		}
	}

	return nil, warnings
}

func untagPaths(store *storage.Storage, tx *storage.Tx, paths, tagArgs []string, recursive, followSymlinks bool) (error, warnings) {
	warnings := make(warnings, 0, 10)

	files := make(entities.Files, 0, len(paths))
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err), warnings
		}

		log.Infof(2, "%v: resolving path", path)

		stat, err := os.Lstat(absPath)
		if err != nil {
			switch {
			case os.IsNotExist(err), os.IsPermission(err):
				// ignore
			default:
				return err, nil
			}
		} else if stat.Mode()&os.ModeSymlink != 0 && followSymlinks {
			absPath, err = _path.Dereference(absPath)
			if err != nil {
				return err, nil
			}
		}

		file, err := store.FileByPath(tx, absPath)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve file: %v", path, err), warnings
		}
		if file == nil {
			warnings = append(warnings, fmt.Sprintf("%v: file is not tagged", path))
			continue
		}

		files = append(files, file)

		if recursive {
			childFiles, err := store.FilesByDirectory(tx, file.Path())
			if err != nil {
				return fmt.Errorf("%v: could not retrieve files for directory: %v", file.Path(), err), warnings
			}

			files = append(files, childFiles...)
		}
	}

	for _, tagArg := range tagArgs {
		tagName, valueName := parseTagEqValueName(tagArg)

		tag, err := store.TagByName(tx, tagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err), warnings
		}
		if tag == nil {
			warnings = append(warnings, fmt.Sprintf("no such tag '%v'", tagName))
			continue
		}

		value, err := store.ValueByName(tx, valueName)
		if err != nil {
			return fmt.Errorf("could not retrieve value '%v': %v", valueName, err), warnings
		}
		if value == nil {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", valueName))
			continue
		}

		for _, file := range files {
			if err := store.DeleteFileTag(tx, file.Id, tag.Id, value.Id); err != nil {
				switch err.(type) {
				case storage.FileTagDoesNotExist:
					exists, err := store.FileTagExists(tx, file.Id, tag.Id, value.Id, false)
					if err != nil {
						return fmt.Errorf("could not check if tag exists: %v", err), warnings
					}

					if exists {
						if value.Id != 0 {
							warnings = append(warnings, fmt.Sprintf("%v: cannot remove '%v=%v': delete implication  to remove this tag.", file.Path(), tag.Name, value.Name))
						} else {
							warnings = append(warnings, fmt.Sprintf("%v: cannot remove '%v': delete implication to remove this tag.", file.Path(), tag.Name))
						}
					} else {
						if value.Id != 0 {
							warnings = append(warnings, fmt.Sprintf("%v: file is not tagged '%v=%v'.", file.Path(), tag.Name, value.Name))
						} else {
							warnings = append(warnings, fmt.Sprintf("%v: file is not tagged '%v'.", file.Path(), tag.Name))
						}
					}
				default:
					return fmt.Errorf("%v: could not remove tag '%v', value '%v': %v", file.Path(), tag.Name, value.Name, err), warnings
				}
			}
		}
	}

	return nil, warnings
}
