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
	"os"
	"path/filepath"
	"strings"
	"time"
	"tmsu/common/fingerprint"
	"tmsu/common/log"
	_path "tmsu/common/path"
	"tmsu/entities"
	"tmsu/storage"
)

var RepairCommand = Command{
	Name:     "repair",
	Aliases:  []string{"fix"},
	Synopsis: "Repair the database",
	Description: `tmsu [OPTION]... repair [PATH]...
tmsu [OPTION]... repair --manual OLD NEW

Fixes broken paths, stale fingerprints in the database caused by file modifications and moves.

Modified files are identified by a change to the file's modification time or file size. These files are repaired by updating the details in the database.

An attempt is made to find missing files under PATHs specified. If a file with the same fingerprint is found then the database is updated with the new file's details. If no PATHs are specified, or no match can be found, then the file is instead reported as missing.

Files that have been both moved and modified cannot be repaired and must be manually relocated.

When run with the --manual option, any paths that begin with OLD are updated to begin with NEW. Any affected files' fingerprints are updated providing the file exists at the new location. No further repairs are attempted in this mode.

Examples:

    $ tmsu repair                                
    $ tmsu repair /new/path                      # look for missing files here
    $ tmsu repair --path /home/sally             # repair subset of database
    $ tmsu repair --manual /home/bob /home/fred  # manually repair paths`,
	Options: Options{{"--path", "-p", "limit repair to files in database under path", true, ""},
		{"--pretend", "-P", "do not make any changes", false, ""},
		{"--remove", "-R", "remove missing files from the database", false, ""},
		{"--manual", "-m", "manually relocate files", false, ""},
		{"--unmodified", "-u", "recalculate fingerprints for unmodified files", false, ""}},
	Exec: repairExec,
}

// unexported

func repairExec(options Options, args []string) error {
	pretend := options.HasOption("--pretend")

	if options.HasOption("--manual") {
		fromPath := args[0]
		toPath := args[1]

		if err := manualRepair(fromPath, toPath, pretend); err != nil {
			return err
		}
	} else {
		searchPaths := args
		limitPath := "/" //TODO Windows
		removeMissing := options.HasOption("--remove")
		recalcUnmodified := options.HasOption("--unmodified")
		if options.HasOption("--path") {
			limitPath = options.Get("--path").Argument
		}

		if err := fullRepair(searchPaths, limitPath, removeMissing, recalcUnmodified, pretend); err != nil {
			return err
		}
	}

	return nil
}

func manualRepair(fromPath, toPath string, pretend bool) error {
	absFromPath, err := filepath.Abs(fromPath)
	if err != nil {
		return fmt.Errorf("%v: could not determine absolute path", err)
	}

	absToPath, err := filepath.Abs(toPath)
	if err != nil {
		return fmt.Errorf("%v: could not determine absolute path", err)
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if err := store.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer store.Commit()

	log.Infof(2, "retrieving files under '%v' from the database", fromPath)

	dbFile, err := store.FileByPath(absFromPath)
	if err != nil {
		return fmt.Errorf("%v: could not retrieve file: %v", fromPath, err)
	}

	if dbFile != nil {
		log.Infof(2, "%v: updating to %v", fromPath, toPath)

		if !pretend {
			if err := manualRepairFile(store, dbFile, absToPath); err != nil {
				return err
			}
		}
	}

	dbFiles, err := store.FilesByDirectory(absFromPath)
	if err != nil {
		return fmt.Errorf("could not retrieve files from storage: %v", err)
	}

	for _, dbFile = range dbFiles {
		relFileFromPath := _path.Rel(dbFile.Path())
		absFileToPath := strings.Replace(dbFile.Path(), absFromPath, absToPath, 1)
		relFileToPath := _path.Rel(absFileToPath)

		log.Infof(2, "%v: updating to %v", relFileFromPath, relFileToPath)

		if !pretend {
			if err := manualRepairFile(store, dbFile, absFileToPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func manualRepairFile(store *storage.Storage, file *entities.File, toPath string) error {
	var fingerprint fingerprint.Fingerprint
	var modTime time.Time
	var size int64
	var isDir bool

	stat, err := os.Stat(toPath)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return fmt.Errorf("%v: permission denied", toPath)
		case os.IsNotExist(err):
			return fmt.Errorf("%v: file not found", toPath)
		default:
			return err
		}

		modTime = file.ModTime
		size = file.Size
		isDir = file.IsDir
	} else {
		modTime = stat.ModTime()
		size = stat.Size()
		isDir = stat.IsDir()
	}

	_, err = store.UpdateFile(file.Id, toPath, fingerprint, modTime, size, isDir)

	return err
}

func fullRepair(searchPaths []string, limitPath string, removeMissing, recalcUnmodified, pretend bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if err := store.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer store.Commit()

	fingerprintAlgorithm, err := store.SettingAsString("fingerprintAlgorithm")
	if err != nil {
		return err
	}

	log.Infof(2, "retrieving all files from the database")

	dbFiles, err := store.FilesByDirectory(limitPath)
	if err != nil {
		return fmt.Errorf("could not retrieve files from storage: %v", err)
	}

	log.Infof(2, "retrieved %v files from the database", len(dbFiles))

	unmodfied, modified, missing := determineStatuses(dbFiles)

	if recalcUnmodified {
		if err = repairUnmodified(store, unmodfied, pretend, fingerprintAlgorithm); err != nil {
			return err
		}
	}

	if err = repairModified(store, modified, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMoved(store, missing, searchPaths, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMissing(store, missing, pretend, removeMissing); err != nil {
		return err
	}

	if err = deleteUntaggedFiles(store, dbFiles); err != nil {
		return err
	}

	if err = deleteUnusedValues(store); err != nil {
		return err
	}

	return nil
}

func deleteUntaggedFiles(store *storage.Storage, files entities.Files) error {
	log.Infof(2, "purging untagged files")

	fileIds := make([]entities.FileId, len(files))
	for index, file := range files {
		fileIds[index] = file.Id
	}

	return store.DeleteUntaggedFiles(fileIds)
}

func deleteUnusedValues(store *storage.Storage) error {
	log.Infof(2, "purging unused values")

	values, err := store.Values()
	if err != nil {
		return fmt.Errorf("could not retrieve set of values")
	}

	valueIds := make([]entities.ValueId, len(values))
	for index, value := range values {
		valueIds[index] = value.Id
	}

	return store.DeleteUnusedValues(valueIds)
}

func determineStatuses(dbFiles entities.Files) (unmodified, modified, missing entities.Files) {
	log.Infof(2, "determining file statuses")

	unmodified = make(entities.Files, 0, 10)
	modified = make(entities.Files, 0, 10)
	missing = make(entities.Files, 0, 10)

	for _, dbFile := range dbFiles {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("%v: permission denied", dbFile.Path())
				continue
			case os.IsNotExist(err):
				log.Infof(2, "%v: missing", dbFile.Path())
				missing = append(missing, dbFile)
				continue
			}
		}

		if dbFile.ModTime == stat.ModTime().UTC() && dbFile.Size == stat.Size() {
			log.Infof(2, "%v: unmodified", dbFile.Path())
			unmodified = append(unmodified, dbFile)
		} else {
			log.Infof(2, "%v: modified", dbFile.Path())
			modified = append(modified, dbFile)
		}
	}

	return
}

func repairUnmodified(store *storage.Storage, unmodified entities.Files, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "recalculating fingerprints for unmodified files")

	for _, dbFile := range unmodified {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			return err
		}

		fingerprint, err := fingerprint.Create(dbFile.Path(), fingerprintAlgorithm)
		if err != nil {
			log.Warnf("%v: could not create fingerprint: %v", dbFile.Path(), err)
			continue
		}

		if !pretend {
			_, err := store.UpdateFile(dbFile.Id, dbFile.Path(), fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", dbFile.Path(), err)
			}
		}

		fmt.Printf("%v: recalculated fingerprint\n", dbFile.Path())
	}

	return nil
}

func repairModified(store *storage.Storage, modified entities.Files, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing modified files")

	for _, dbFile := range modified {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			return err
		}

		fingerprint, err := fingerprint.Create(dbFile.Path(), fingerprintAlgorithm)
		if err != nil {
			log.Warnf("%v: could not create fingerprint: %v", dbFile.Path(), err)
			continue
		}

		if !pretend {
			_, err := store.UpdateFile(dbFile.Id, dbFile.Path(), fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", dbFile.Path(), err)
			}

		}

		fmt.Printf("%v: updated fingerprint\n", dbFile.Path())
	}

	return nil
}

func repairMoved(store *storage.Storage, missing entities.Files, searchPaths []string, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing moved files")

	if len(missing) == 0 || len(searchPaths) == 0 {
		// don't bother enumerating filesystem if nothing to do
		return nil
	}

	pathsBySize, err := buildPathBySizeMap(searchPaths)
	if err != nil {
		return err
	}

	for index, dbFile := range missing {
		log.Infof(2, "%v: searching for new location", dbFile.Path())

		pathsOfSize := pathsBySize[dbFile.Size]
		log.Infof(2, "%v: file is of size %v, identified %v files of this size", dbFile.Path(), dbFile.Size, len(pathsOfSize))

		for _, candidatePath := range pathsOfSize {
			candidateFile, err := store.FileByPath(candidatePath)
			if err != nil {
				return err
			}
			if candidateFile != nil {
				// file is already tagged
				continue
			}

			stat, err := os.Stat(candidatePath)
			if err != nil {
				return fmt.Errorf("%v: could not stat file: %v", candidatePath, err)
			}

			fingerprint, err := fingerprint.Create(candidatePath, fingerprintAlgorithm)
			if err != nil {
				return fmt.Errorf("%v: could not create fingerprint: %v", candidatePath, err)
			}

			if fingerprint == dbFile.Fingerprint {
				if !pretend {
					_, err := store.UpdateFile(dbFile.Id, candidatePath, dbFile.Fingerprint, stat.ModTime(), dbFile.Size, dbFile.IsDir)
					if err != nil {
						return fmt.Errorf("%v: could not update file in database: %v", dbFile.Path(), err)
					}
				}

				fmt.Printf("%v: updated path to %v\n", dbFile.Path(), candidatePath)

				missing[index] = nil

				break
			}
		}
	}

	return nil
}

func repairMissing(store *storage.Storage, missing entities.Files, pretend, force bool) error {
	for _, dbFile := range missing {
		if dbFile == nil {
			continue
		}

		if force {
			if !pretend {
				if err := store.DeleteFileTagsByFileId(dbFile.Id); err != nil {
					return fmt.Errorf("%v: could not delete file-tags: %v", dbFile.Path(), err)
				}
			}

			fmt.Printf("%v: removed\n", dbFile.Path())
		} else {
			fmt.Printf("%v: missing\n", dbFile.Path())
		}
	}

	return nil
}

func buildPathBySizeMap(paths []string) (map[int64][]string, error) {
	log.Infof(2, "building map of paths by size")

	pathsBySize := make(map[int64][]string, 10)

	for _, path := range paths {
		if err := buildPathBySizeMapRecursive(path, pathsBySize); err != nil {
			return nil, err
		}
	}

	log.Infof(2, "path by size map has %v sizes", len(pathsBySize))

	return pathsBySize, nil
}

func buildPathBySizeMapRecursive(path string, pathBySizeMap map[int64][]string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path", path)
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		switch {
		case os.IsPermission(err):
			log.Warnf("%v: permission denied", path)
		default:
			return err
		}
	}

	if stat.IsDir() {
		log.Infof(3, "%v: examining directory contents", absPath)

		dir, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("%v: could not open directory: %v", path, err)
		}

		names, err := dir.Readdirnames(0)
		dir.Close()
		if err != nil {
			return fmt.Errorf("%v: could not read directory entries: %v", path, err)
		}

		for _, name := range names {
			childPath := filepath.Join(path, name)
			if err := buildPathBySizeMapRecursive(childPath, pathBySizeMap); err != nil {
				return err
			}
		}
	} else {
		log.Infof(3, "%v: file is of size %v", absPath, stat.Size())

		filesOfSize, ok := pathBySizeMap[stat.Size()]
		if ok {
			pathBySizeMap[stat.Size()] = append(filesOfSize, absPath)
		} else {
			pathBySizeMap[stat.Size()] = []string{absPath}
		}
	}

	return nil
}
