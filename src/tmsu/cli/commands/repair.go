/*
Copyright 2011-2012 Paul Ruane.

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
	"os"
	"path/filepath"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

//TODO add missing implicit taggings
//TODO add missing fingerprints, mod_times and sizes
//TODO delete implicitly tagged files that are missing
//TODO handle directory being replaced by a file (currently causes error)

type RepairCommand struct {
	verbose bool
}

func (RepairCommand) Name() cli.CommandName {
	return "repair"
}

func (RepairCommand) Synopsis() string {
	return "Repair the database"
}

func (RepairCommand) Description() string {
	return `tmsu repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

Repairs tagged files and directories under PATHs by:

    1. Identifying modified files.
    2. Identifying new files.
    3. Identifying moved files.

Where no PATHS are specified all tagged files and directories fingerprints in
the database are checked and their fingerprints updated where modifications are
found.

Modified files are identified by a change of the modification timestamp.

New files that lie under a tagged directory (and thus are implicitly tagged)
are added to the database.

Moved files are identified by file fingerprint and will only be found if they
have been moved under one of the specified PATHs. (As such, moved files cannot
be identified where no PATHs are specified.)`
}

func (RepairCommand) Options() cli.Options {
	return cli.Options{}
}

func (command RepairCommand) Exec(options cli.Options, args []string) error {
	command.verbose = cli.HasOption(options, "--verbose")

	pathsBySize, err := command.buildFileSystemMap(args)
	if err != nil {
		return err
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	for _, path := range args {
		entries, err := store.FilesByDirectory(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			err := command.checkEntry(entry, store, pathsBySize)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (command RepairCommand) checkEntry(entry *database.File, store *storage.Storage, pathsBySize map[int64][]string) error {
	if command.verbose {
		fmt.Printf("Checking '%v'.\n", entry.Path())
	}

	info, err := os.Stat(entry.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			err = command.processMissingEntry(entry, pathsBySize, store)
			if err != nil {
				return err
			}
		case os.IsPermission(err):
			log.Warnf("'%v': Permission denied.", entry.Path())
		default:
			log.Warnf("'%v': %v", entry.Path(), err)
		}

		return nil
	}
	modTime := info.ModTime().UTC()
	size := info.Size()

	if info.IsDir() {
		command.processDirectory(store, entry.Path())
	}

	if modTime.Unix() != entry.ModTimestamp.Unix() || size != entry.Size {
		if command.verbose {
			fmt.Printf("'%v' is modified: updating entry in database.\n", entry.Path())
		}

		fingerprint, err := fingerprint.Create(entry.Path())
		if err != nil {
			return err
		}

		if fingerprint != entry.Fingerprint {
			err = store.UpdateFile(entry.Id, entry.Path(), fingerprint, modTime, size)
			if err != nil {
				return err
			}

			fmt.Printf("'%v': Modified.\n", entry.Path())
		}
	}

	if command.verbose {
		fmt.Printf("'%v': Unchanged.\n", entry.Path())
	}

	return nil
}

func (command RepairCommand) processDirectory(store *storage.Storage, path string) error {
	if command.verbose {
		fmt.Printf("Checking directory contents.\n")
	}

	dir, err := os.Open(path)
	if err != nil {
		return err
	}

	filenames, err := dir.Readdirnames(0)
	if err != nil {
		dir.Close()
		return err
	}

	dir.Close()

	for _, filename := range filenames {
		childPath := filepath.Join(path, filename)

		file, err := store.FileByPath(childPath)
		if err != nil {
			return err
		}
		if file == nil {
			if command.verbose {
				fmt.Printf("'%v' is new: adding to database.\n", childPath)
			}

			cli.AddFile(store, childPath)
			//TODO add implicit tags
		}
	}

	//TODO remove files in the directory that are no longer on disk

	return nil
}

func (command RepairCommand) processMissingEntry(entry *database.File, pathsBySize map[int64][]string, store *storage.Storage) error {
	if command.verbose {
		fmt.Printf("Missing: searching for new location.\n")
	}

	paths, found := pathsBySize[entry.Size]
	if found {
		for _, path := range paths {
			fingerprint, err := fingerprint.Create(path)
			if err != nil {
				return err
			}

			if fingerprint == "" {
				break
			}

			if fingerprint == entry.Fingerprint {
				if command.verbose {
					fmt.Printf("'%v' found at '%v': updating location in database.\n", entry.Path(), path)
				}

				info, err := os.Stat(path)
				if err != nil {
					return err
				}

				err = store.UpdateFile(entry.Id, path, entry.Fingerprint, info.ModTime().UTC(), info.Size())
				if err != nil {
					return err
				}

				fmt.Printf("'%v': Moved to '%v'.\n", entry.Path(), path)
				return nil
			}
		}
	}

	log.Warnf("'%v': Missing.", entry.Path())
	return nil
}

func (command RepairCommand) buildFileSystemMap(paths []string) (map[int64][]string, error) {
	if command.verbose {
		fmt.Printf("Building map of files by size.\n")
	}

	pathsBySize := make(map[int64][]string)

	for _, path := range paths {
		err := command.buildFileSystemMapRecursive(path, pathsBySize)
		if err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("'%v': Permission denied.")
				continue
			}

			return nil, err
		}
	}

	if command.verbose {
		fmt.Printf("Finished building map of files by size.\n")
	}

	return pathsBySize, nil
}

func (command RepairCommand) buildFileSystemMapRecursive(path string, pathsBySize map[int64][]string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("3")
		return err
	}

	if info.IsDir() {
		dirEntries, err := file.Readdir(0)
		if err != nil {
			return err
		}
		file.Close()

		for _, dirEntry := range dirEntries {
			dirEntryPath := filepath.Join(path, dirEntry.Name())
			command.buildFileSystemMapRecursive(dirEntryPath, pathsBySize)
		}
	} else {
		file.Close()

		if info.Size() > 0 {
			paths, found := pathsBySize[info.Size()]
			if !found {
				paths = make([]string, 0, 10)
			}
			paths = append(paths, path)
			pathsBySize[info.Size()] = paths
		}
	}

	return nil
}
