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
	"tmsu/common/fingerprint"
	"tmsu/common/log"
	_path "tmsu/common/path"
	"tmsu/entities"
	"tmsu/storage"
)

var RepairCommand = Command{
	Name:     "repair",
	Synopsis: "Repair the database",
	Description: `tmsu [OPTION]... repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

Where no PATHS are specified all files in the database are checked.

                                          Reported Repaired
    Modified files                          yes      yes
    Moved files                             yes      yes
    Missing files                           yes      no
    Untagged files                          yes
    Unmodified files                        no       

Modified files are identified by a change to the file's modification time or
file size. These files are repaired by updating the modification time, size and
fingerprint in the database.

Moved files will only be repaired if a file with the same fingerprint can be
found under PATHs. Files that have been both moved and modified will not be
identified.

Missing files are reported but are not, by default, removed from the database
as this would destroy the tagging information associated with them. To remove
these files use the --remove option.

Examples:

    $ tmsu repair
    $ tmsu repair /some/path /new/path
    $ tmsu repair --remove`,
	Options: Options{{"--pretend", "-P", "do not make any changes", false, ""},
		{"--remove", "-R", "remove missing files from the database", false, ""},
		{"--unmodified", "-u", "recalculate fingerprints for unmodified files", false, ""}},
	Exec: repairExec,
}

func repairExec(options Options, args []string) error {
	pretend := options.HasOption("--pretend")
	removeMissing := options.HasOption("--remove")
	recalcUnmodified := options.HasOption("--unmodified")

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

	if len(args) == 0 {
		return repairDatabase(store, pretend, removeMissing, recalcUnmodified, fingerprintAlgorithm)
	}

	return repairPaths(store, args, pretend, removeMissing, recalcUnmodified, fingerprintAlgorithm)
}

//- unexported

func repairDatabase(store *storage.Storage, pretend, removeMissing, recalcUnmodified bool, fingerprintAlgorithm string) error {
	log.Infof(2, "retrieving all files from the database.")

	files, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve paths from storage: %v", err)
	}

	paths := make([]string, len(files))
	for index := 0; index < len(files); index++ {
		paths[index] = files[index].Path()
	}

	err = repairFiles(store, paths, pretend, removeMissing, recalcUnmodified, fingerprintAlgorithm)
	if err != nil {
		return err
	}

	return nil
}

func repairPaths(store *storage.Storage, paths []string, pretend, removeMissing, recalcUnmodified bool, fingerprintAlgorithm string) error {
	absPaths := make([]string, len(paths))

	for index, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		absPaths[index] = absPath
	}

	log.Infof(2, "identifying top-level paths.")

	err := repairFiles(store, absPaths, pretend, removeMissing, recalcUnmodified, fingerprintAlgorithm)
	if err != nil {
		return err
	}

	return nil
}

func repairFiles(store *storage.Storage, paths []string, pretend, removeMissing, recalcUnmodified bool, fingerprintAlgorithm string) error {
	tree := _path.NewTree()
	for _, path := range paths {
		tree.Add(path, false)
	}
	paths = tree.TopLevel().Paths()

	fsPaths, err := enumerateFileSystemPaths(paths)
	if err != nil {
		return err
	}

	dbPaths, err := enumerateDatabasePaths(store, paths)
	if err != nil {
		return err
	}

	unmodfied, untagged, modified, missing := determineStatuses(fsPaths, dbPaths)

	if recalcUnmodified {
		if err = repairUnmodified(store, unmodfied, pretend, fingerprintAlgorithm); err != nil {
			return err
		}
	}

	if err = repairModified(store, modified, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMoved(store, missing, untagged, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMissing(store, missing, pretend, removeMissing); err != nil {
		return err
	}

	for path := range untagged {
		log.Infof(1, "%v: untagged", path)
	}

	//TODO cleanup: any files that have no tags: remove
	//TODO cleanup: any tags that do not correspond to a file: remove

	return nil
}

type fileInfoMap map[string]os.FileInfo
type fileIdAndInfoMap map[string]struct {
	fileId entities.FileId
	stat   os.FileInfo
}
type databaseFileMap map[string]entities.File

func determineStatuses(fsPaths fileInfoMap, dbPaths databaseFileMap) (unmodified fileIdAndInfoMap, untagged fileInfoMap, modified fileIdAndInfoMap, missing databaseFileMap) {
	log.Infof(2, "determining file statuses")

	unmodified = make(fileIdAndInfoMap, 100)
	untagged = make(fileInfoMap, 100)
	modified = make(fileIdAndInfoMap, 100)
	missing = make(databaseFileMap, 100)

	for path, stat := range fsPaths {
		if dbFile, isTagged := dbPaths[path]; isTagged {
			dbFileAndStat := struct {
				fileId entities.FileId
				stat   os.FileInfo
			}{dbFile.Id, stat}

			if dbFile.ModTime == stat.ModTime().UTC() && dbFile.Size == stat.Size() {
				unmodified[path] = dbFileAndStat
			} else {
				modified[path] = dbFileAndStat
			}
		} else {
			untagged[path] = stat
		}
	}

	for path, dbFile := range dbPaths {
		if _, found := fsPaths[path]; !found {
			missing[path] = dbFile
		}
	}

	return
}

func repairUnmodified(store *storage.Storage, unmodified fileIdAndInfoMap, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "recalculating fingerprints for unmodified files")

	for path, fileIdAndStat := range unmodified {
		fileId := fileIdAndStat.fileId
		stat := fileIdAndStat.stat

		log.Infof(1, "%v: unmodified", path)

		fingerprint, err := fingerprint.Create(path, fingerprintAlgorithm)
		if err != nil {
			return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
		}

		if !pretend {
			_, err := store.UpdateFile(fileId, path, fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", path, err)
			}
		}

	}

	return nil
}

func repairModified(store *storage.Storage, modified fileIdAndInfoMap, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing modified files")

	for path, fileIdAndStat := range modified {
		fileId := fileIdAndStat.fileId
		stat := fileIdAndStat.stat

		log.Infof(1, "%v: modified", path)

		fingerprint, err := fingerprint.Create(path, fingerprintAlgorithm)
		if err != nil {
			return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
		}

		if !pretend {
			_, err := store.UpdateFile(fileId, path, fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", path, err)
			}
		}

	}

	return nil
}

func repairMoved(store *storage.Storage, missing databaseFileMap, untagged fileInfoMap, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing moved files")

	moved := make([]string, 0, 10)
	for path, dbFile := range missing {
		log.Infof(2, "%v: searching for new location", path)

		for candidatePath, stat := range untagged {
			if stat.Size() == dbFile.Size {
				fingerprint, err := fingerprint.Create(candidatePath, fingerprintAlgorithm)
				if err != nil {
					return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
				}

				if fingerprint == dbFile.Fingerprint {
					log.Infof(1, "%v: moved to %v", path, candidatePath)

					moved = append(moved, path)

					if !pretend {
						_, err := store.UpdateFile(dbFile.Id, candidatePath, dbFile.Fingerprint, stat.ModTime(), dbFile.Size, dbFile.IsDir)
						if err != nil {
							return fmt.Errorf("%v: could not update file in database: %v", path, err)
						}
					}

					delete(untagged, candidatePath)

					break
				}
			}
		}
	}

	for _, path := range moved {
		delete(missing, path)
	}

	return nil
}

func repairMissing(store *storage.Storage, missing databaseFileMap, pretend, force bool) error {
	for path, dbFile := range missing {
		if force && !pretend {
			if err := store.DeleteFileTagsByFileId(dbFile.Id); err != nil {
				return fmt.Errorf("%v: could not delete file-tags: %v", path, err)
			}

			log.Infof(1, "%v: removed", path)
		} else {
			log.Infof(1, "%v: missing", path)
		}
	}

	return nil
}

func enumerateFileSystemPaths(paths []string) (fileInfoMap, error) {
	files := make(fileInfoMap, 100)

	for _, path := range paths {
		if err := enumerateFileSystemPath(files, path); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func enumerateFileSystemPath(files fileInfoMap, path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			log.Warnf("%v: permission denied", path)
			return nil
		case os.IsNotExist(err):
			return nil
		default:
			return fmt.Errorf("%v: could not stat file: %v", path, err)
		}
	}

	files[path] = stat

	if stat.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("%v: could not open file: %v", path, err)
		}

		childFilenames, err := file.Readdirnames(0)
		file.Close()
		if err != nil {
			return fmt.Errorf("%v: could not read directory: %v", file.Name(), err)
		}

		for _, childFilename := range childFilenames {
			childPath := filepath.Join(path, childFilename)
			enumerateFileSystemPath(files, childPath)
		}
	}

	return nil
}

func enumerateDatabasePaths(store *storage.Storage, paths []string) (databaseFileMap, error) {
	dbFiles := make(databaseFileMap, 100)

	for _, path := range paths {
		file, err := store.FileByPath(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not retrieve file from database: %v", path, err)
		}
		if file != nil {
			dbFiles[file.Path()] = *file
		}

		files, err := store.FilesByDirectory(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not retrieve files from database: %v", path, err)
		}

		for _, file = range files {
			dbFiles[file.Path()] = *file
		}
	}

	return dbFiles, nil
}
