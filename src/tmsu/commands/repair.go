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
	"tmsu/common"
	"tmsu/database"
	"tmsu/fingerprint"
)

type RepairCommand struct{}

func (RepairCommand) Name() string {
	return "repair"
}

func (RepairCommand) Synopsis() string {
	return "Repair the database"
}

func (RepairCommand) Description() string {
	return `tmsu repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

Repairs tagged files and directories under PATHS by:

    1. Updating the stored fingerprints of modified files and directories.
    2. Updating the path of missing files and directories where an untagged file
       or directory with the same fingerprint can be found in PATHs.

Where no PATHS are specified all tagged files and directories fingerprints in
the database are checked and their fingerprints updated where modifications are
found. (In this mode file move repairs are not performed.)`
}

func (command RepairCommand) Exec(args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}

	pathsByFingerprint, err := command.buildFileSystemFingerprints(args)
	if err != nil {
		return err
	}

	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, path := range args {
		childEntries, err := db.FilesByDirectory(path)
		if err != nil {
			return err
		}

		for _, childEntry := range childEntries {
			err := command.checkEntry(childEntry, db, pathsByFingerprint)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (command RepairCommand) checkEntry(entry *database.File, db *database.Database, pathsByFingerprint map[fingerprint.Fingerprint][]string) error {
	info, err := os.Stat(entry.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			err = command.processMissingEntry(entry, pathsByFingerprint, db)
			if err != nil {
				return err
			}
		case os.IsPermission(err):
			common.Warnf("'%v': Permission denied.", entry.Path())
		default:
			common.Warn("'%v': %v", err)
		}

		return nil
	}
	modTime := info.ModTime().UTC()

	fingerprint, err := fingerprint.Create(entry.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			common.Warnf("'%v': Missing", entry.Path())
		case os.IsPermission(err):
			common.Warnf("'%v': Permission denied", entry.Path())
		default:
			common.Warn("'%v': %v", err)
		}

		return nil
	}

	if info.IsDir() {
		command.processDirectory(db, entry.Path())
	}

	if modTime.Unix() != entry.ModTimestamp.Unix() || fingerprint != entry.Fingerprint {
		err := db.UpdateFile(entry.Id, entry.Path(), fingerprint, modTime)
		if err != nil {
			return err
		}

		fmt.Printf("'%v': Repaired.\n", entry.Path())
	}

	return nil
}

func (command RepairCommand) processDirectory(db *database.Database, path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	filenames, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		childPath := filepath.Join(path, filename)

		file, err := db.FileByPath(childPath)
		if err != nil {
			return err
		}
		if file == nil {
			addFile(db, childPath)
		}
	}

	return nil
}

func (command RepairCommand) processMissingEntry(entry *database.File, pathsByFingerprint map[fingerprint.Fingerprint][]string, db *database.Database) error {
	paths, ok := pathsByFingerprint[entry.Fingerprint]
	if !ok {
		common.Warnf("'%v': Missing.", entry.Path())
		return nil
	}

	switch len(paths) {
	case 0:
		panic("No paths for fingerprint.")
	case 1:
		newPath, err := filepath.Abs(paths[0])
		if err != nil {
			return err
		}

		info, err := os.Stat(newPath)
		if err != nil {
			return err
		}

		db.UpdateFile(entry.Id, newPath, entry.Fingerprint, info.ModTime().UTC())

		fmt.Printf("'%v': Repaired (moved to '%v').\n", entry.Path(), newPath)
	default:
		common.Warnf("'%v': Cannot repair: file moved to multiple destinations.", entry.Path())
	}

	return nil
}

func (command RepairCommand) buildFileSystemFingerprints(paths []string) (map[fingerprint.Fingerprint][]string, error) {
	pathsByFingerprint := make(map[fingerprint.Fingerprint][]string)

	for _, path := range paths {
		err := command.buildFileSystemFingerprintsRecursive(path, pathsByFingerprint)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				continue
			case os.IsPermission(err):
				continue
			}

			return nil, err
		}
	}

	return pathsByFingerprint, nil
}

func (command RepairCommand) buildFileSystemFingerprintsRecursive(path string, pathsByFingerprint map[fingerprint.Fingerprint][]string) error {
	path = filepath.Clean(path)

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return err
	}

	paths, ok := pathsByFingerprint[fingerprint]
	if !ok {
		paths = make([]string, 0, 10)
	}
	paths = append(paths, path)
	pathsByFingerprint[fingerprint] = paths

	if common.IsDir(path) {
		dir, err := os.Open(path)
		if err != nil {
			return err
		}

		dirEntries, err := dir.Readdir(0)
		if err != nil {
			return err
		}

		for _, dirEntry := range dirEntries {
			dirEntryPath := filepath.Join(path, dirEntry.Name())
			command.buildFileSystemFingerprintsRecursive(dirEntryPath, pathsByFingerprint)
		}
	}

	return nil
}
