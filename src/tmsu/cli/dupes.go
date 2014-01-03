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
	"tmsu/common/fingerprint"
	"tmsu/common/log"
	_path "tmsu/common/path"
	"tmsu/entities"
	"tmsu/storage"
)

var DupesCommand = Command{
	Name:     "dupes",
	Synopsis: "Identify duplicate files",
	Description: `tmsu dupes [FILE]...

Identifies all files in the database that are exact duplicates of FILE. If no
FILE is specified then identifies duplicates between files in the database.

Examples:

    $ tmsu dupes
    Set of 2 duplicates:
      /tmp/song.mp3
      /tmp/copy of song.mp3
    $ tmsu dupes /tmp/song.mp3
    /tmp/copy of song.mp3`,
	Options: Options{Option{"--recursive", "-r", "recursively check directory contents", false, ""}},
	Exec:    dupesExec,
}

func dupesExec(options Options, args []string) error {
	recursive := options.HasOption("--recursive")

	switch len(args) {
	case 0:
		findDuplicatesInDb()
	default:
		return findDuplicatesOf(args, recursive)
	}

	return nil
}

func findDuplicatesInDb() error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "identifying duplicate files.")

	fileSets, err := store.DuplicateFiles()
	if err != nil {
		return fmt.Errorf("could not identify duplicate files: %v", err)
	}

	log.Infof(2, "found %v sets of duplicate files.", len(fileSets))

	for index, fileSet := range fileSets {
		if index > 0 {
			fmt.Println()
		}

		fmt.Printf("Set of %v duplicates:\n", len(fileSet))

		for _, file := range fileSet {
			relPath := _path.Rel(file.Path())
			fmt.Printf("  %v\n", relPath)
		}
	}

	return nil
}

func findDuplicatesOf(paths []string, recursive bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if recursive {
		p, err := _path.Enumerate(paths)
		if err != nil {
			return fmt.Errorf("could not enumerate paths: %v", err)
		}

		paths = make([]string, len(p))
		for index, path := range p {
			paths[index] = path.Path
		}
	}

	first := true
	for _, path := range paths {
		log.Infof(2, "%v: identifying duplicate files.", path)

		fp, err := fingerprint.Create(path)
		if err != nil {
			return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
		}

		if fp == fingerprint.Fingerprint("") {
			continue
		}

		files, err := store.FilesByFingerprint(fp)
		if err != nil {
			return fmt.Errorf("%v: could not retrieve files matching fingerprint '%v': %v", path, fp, err)
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not determine absolute path: %v", path, err)
		}

		// filter out the file we're searching on
		dupes := files.Where(func(file *entities.File) bool { return file.Path() != absPath })

		if len(paths) > 1 && len(dupes) > 0 {
			if first {
				first = false
			} else {
				fmt.Println()
			}

			fmt.Printf("%v:\n", path)

			for _, dupe := range dupes {
				relPath := _path.Rel(dupe.Path())
				fmt.Printf("  %v\n", relPath)
			}
		} else {
			for _, dupe := range dupes {
				relPath := _path.Rel(dupe.Path())
				fmt.Println(relPath)
			}
		}
	}

	return nil
}
