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

package commands

import (
	"fmt"
	"path/filepath"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	_path "tmsu/path"
	"tmsu/storage"
	"tmsu/storage/database"
)

type DupesCommand struct {
	verbose   bool
	recursive bool
}

func (DupesCommand) Name() cli.CommandName {
	return "dupes"
}

func (DupesCommand) Synopsis() string {
	return "Identify duplicate files"
}

func (DupesCommand) Description() string {
	return `tmsu dupes [FILE]...

Identifies all files in the database that are exact duplicates of FILE. If no
FILE is specified then identifies duplicates between files in the database.`
}

func (DupesCommand) Options() cli.Options {
	return cli.Options{cli.Option{"--recursive", "-r", "recursively check directory contents", false, ""}}
}

func (command DupesCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.recursive = options.HasOption("--recursive")

	switch len(args) {
	case 0:
		command.findDuplicatesInDb()
	default:
		return command.findDuplicatesOf(args)
	}

	return nil
}

func (command DupesCommand) findDuplicatesInDb() error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.verbose {
		log.Info("identifying duplicate files.")
	}

	fileSets, err := store.DuplicateFiles()
	if err != nil {
		return fmt.Errorf("could not identify duplicate files: %v", err)
	}

	if command.verbose {
		log.Infof("found %v sets of duplicate files.", len(fileSets))
	}

	for index, fileSet := range fileSets {
		if index > 0 {
			log.Print()
		}

		log.Printf("Set of %v duplicates:", len(fileSet))

		for _, file := range fileSet {
			relPath := _path.Rel(file.Path())
			log.Printf("  %v", relPath)
		}
	}

	return nil
}

func (command DupesCommand) findDuplicatesOf(paths []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.recursive {
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
		if command.verbose {
			log.Infof("%v: identifying duplicate files.", path)
		}

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
		dupes := files.Where(func(file *database.File) bool { return file.Path() != absPath })

		if len(paths) > 1 && len(dupes) > 0 {
			if first {
				first = false
			} else {
				log.Print()
			}

			log.Printf("%v:", path)

			for _, dupe := range dupes {
				relPath := _path.Rel(dupe.Path())
				log.Printf("  %v", relPath)
			}
		} else {
			for _, dupe := range dupes {
				relPath := _path.Rel(dupe.Path())
				log.Print(relPath)
			}
		}
	}

	return nil
}
