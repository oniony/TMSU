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
the database are repaired. (In this mode file moves cannot be repaired as tmsu
will not know where to look for them.)`
}

func (command RepairCommand) Exec(args []string) error {
    pathByFingerprint, err := command.buildFileSystemFingerprints(args)
    if err != nil {
        return err
    }

    db, err := database.Open()
    if err != nil {
        return err
    }
    defer db.Close()

    for _, path := range args {
        entry, err := db.FileByPath(path)
        if err != nil {
            return err
        }
        if entry != nil {
            err := command.checkEntry(entry, pathsByFingerprint)
            if err != nil {
                return err
            }
        }

        // path might be a directory
        childEntries, err := db.FilesByDirectory(path)
        for _, childEntry := range childEntries {
            err := command.checkEntry(childEntry, pathByFingerprint)
        }
    }
}

func (command RepairCommand) checkEntry(entry *database.File, fileSystemEntries map[fingerprint.Fingerprint]string) error {
    fingerprint, err := fingerprint.Fingerprint(entry.Path())
    if err != nil {
        switch {
        case os.IsNotExist(err):
            //TODO is there an untagged file with same fingerprint?
            common.Warnf(entry.Path(), "Missing - maybe moved")
            return nil
        case os.IsPermission(err):
            common.Warnf("'%v': Permission denied", entry.Path())
            return nil
        default:
            common.Warn("'%v': %v", err)
        }
    }

    if fingerprint != entry.Fingerprint() {
        common.Warn(entry.Path(), "File is modified")
    } else {
        common.Warn(entry.Path(), "File is not modified")
    }

	return nil
}

func (command RepairCommand) buildFileSystemFingerprints(paths []string) (map[fingerprint.Fingerprint]string, error) {
    pathByFingerprint := make(map[fingerprint.Fingerprint]string)

    for _, path := range paths {
        err := command.buildFileSystemFingerprintsRecursive(path, pathByFingerprint)
        if err != nil {
            return nil, err
        }
    }

    return pathByFingerprint, nil
}

func (command RepairCommand) buildFileSystemFingerprintsRecursive(path string, pathByFingerprint map[fingerprint.Fingerprint]string) error {
    path = filepath.Clean(path)

    fingerprint, err := fingerprint.Create(path)
    if err != nil {
        return err
    }

    pathByFingerprint[fingerprint] = path

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
            command.buildFileSystemFingerprintsRecursive(dirEntryPath, pathByFingerprint)
        }
    }

    return nil
}
