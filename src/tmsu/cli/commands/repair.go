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
	"os"
	"path/filepath"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	_path "tmsu/path"
	"tmsu/storage"
	"tmsu/storage/database"
)

type RepairCommand struct {
	verbose bool
	pretend bool
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

Repairs tagged files and directories under PATHs by identifying:

    1. Modified files.
    2. Added files within tagged directories.
    3. Moved files.
    4. Missing files.
    5. Missing implicit taggings.

Where no PATHS are specified all tagged files and directories fingerprints in
the database are checked and their fingerprints updated where modifications are
found.`
}

func (RepairCommand) Options() cli.Options {
	return cli.Options{{"--pretend", "-p", "do not make any changes"}}
}

func (command RepairCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.pretend = options.HasOption("--pretend")

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if len(args) == 0 {
		return command.repairDatabase(store)
	}

	return command.repairPaths(store, args)
}

func (command RepairCommand) repairDatabase(store *storage.Storage) error {
	if command.verbose {
		log.Info("retrieving all files from the database.")
	}

	dbFiles, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve files from storage: %v", err)
	}

	absPaths := make([]string, len(dbFiles))
	for index, file := range dbFiles {
		absPaths[index] = file.Path()
	}

	absPaths, err = _path.Roots(absPaths)
	if err != nil {
		return fmt.Errorf("could not identify top-level paths: %v", err)
	}

	err = command.checkFiles(store, absPaths, dbFiles)
	if err != nil {
		return err
	}

	return nil
}

func (command RepairCommand) repairPaths(store *storage.Storage, paths []string) error {
	absPaths := make([]string, len(paths))
	for index, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
		}

		absPaths[index] = absPath
	}

	absPaths, err := _path.Roots(absPaths)
	if err != nil {
		return fmt.Errorf("could not identify top-level paths: %v", err)
	}

	if command.verbose {
		log.Infof("identifying top-level paths.")
	}

	dbFiles, err := store.FilesByDirectories(absPaths)
	if err != nil {
		return fmt.Errorf("could not retrieve files from database: %v", err)
	}

	err = command.checkFiles(store, absPaths, dbFiles)
	if err != nil {
		return err
	}

	return nil
}

func (command RepairCommand) checkFiles(store *storage.Storage, paths []string, files database.Files) error {
	dbFileByPath := toMap(files)

	fsStatByPath, err := command.buildFileSystemMap(paths)
	if err != nil {
		return fmt.Errorf("could not build file system map: %v", err)
	}

	if command.verbose {
		log.Infof("searching for modified and untagged files")
	}

	untaggedPathsBySize := make(map[int64][]string, 100)
	for path, stat := range fsStatByPath {
		dbFile, found := dbFileByPath[path]
		if found {
			if stat.Size() != dbFile.Size || stat.ModTime().UTC() != dbFile.ModTimestamp {
				if err := command.processModifiedFile(store, dbFile, stat); err != nil {
					return err
				}
			} else {
				if err := command.processTaggedFile(store, dbFile); err != nil {
					return err
				}
			}
		} else {
			if err := command.processUntaggedFile(untaggedPathsBySize, path, stat.Size()); err != nil {
				return err
			}
		}
	}

	if command.verbose {
		log.Infof("searching for missing and removed files")
	}

	for path, dbFile := range dbFileByPath {
		_, found := fsStatByPath[path]
		if !found {
			err := command.processMissingFile(store, untaggedPathsBySize, dbFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (command RepairCommand) processTaggedFile(store *storage.Storage, dbFile *database.File) error {
	//TODO if directory
	//TODO   any file within that is in untagged should be added
	//TODO   apply parent's implicit taggings
	//TODO remember to check 'pretend'
	return nil
}

func (command RepairCommand) processUntaggedFile(untaggedPathsBySize map[int64][]string, path string, size int64) error {
	paths, found := untaggedPathsBySize[size]
	if !found {
		paths = make([]string, 0, 2)
	}

	paths = append(paths, path)
	untaggedPathsBySize[size] = paths

	return nil
}

func (command RepairCommand) processModifiedFile(store *storage.Storage, dbFile *database.File, stat os.FileInfo) error {
	fmt.Printf("'%v': modified.\n", dbFile.Path())

	fingerprint, err := fingerprint.Create(dbFile.Path())
	if err != nil {
		return fmt.Errorf("'%v': could not create fingerprint: %v", dbFile.Path(), err)
	}

	if !command.pretend {
		if _, err := store.UpdateFile(dbFile.Id, dbFile.Path(), fingerprint, stat.ModTime().UTC(), stat.Size()); err != nil {
			fmt.Errorf("'%v': could not update file in database: %v", err)
		}
	}

	return nil
}

func (command RepairCommand) processMissingFile(store *storage.Storage, untaggedPathsBySize map[int64][]string, dbFile *database.File) error {
	if command.verbose {
		log.Infof("'%v': attempting to relocate missing file.", dbFile.Path())
	}

	candidatePaths := untaggedPathsBySize[dbFile.Size]
	for _, candidatePath := range candidatePaths {
		fingerprint, err := fingerprint.Create(candidatePath)
		if err != nil {
			return fmt.Errorf("'%v': could not create fingerprint: %v", candidatePath, err)
		}

		if fingerprint == dbFile.Fingerprint {
			return command.processMovedFile(store, dbFile, candidatePath)
		}
	}

	fmt.Printf("'%v': missing.\n", dbFile.Path())

	if !command.pretend {
		//TODO remove if has no explicit tags
	}

	return nil
}

func (command RepairCommand) processMovedFile(store *storage.Storage, dbFile *database.File, path string) error {
	fmt.Printf("'%v': moved to '%v'.\n", dbFile.Path(), path)

	if !command.pretend {
		if _, err := store.UpdateFile(dbFile.Id, path, dbFile.Fingerprint, dbFile.ModTimestamp, dbFile.Size); err != nil {
			return fmt.Errorf("'%v': could not update file path in database: %v", dbFile.Path(), err)
		}

		//TODO reevalutae implicit file taggings
	}

	return nil
}

func (command RepairCommand) buildFileSystemMap(paths []string) (map[string]os.FileInfo, error) {
	files := make(map[string]os.FileInfo, 100)

	for _, path := range paths {
		if command.verbose {
			log.Infof("'%v': enumerating files", path)
		}

		if err := command.readFiles(files, path); err != nil {
			return nil, fmt.Errorf("'%v': could not read files: %v", path, err)
		}
	}

	return files, nil
}

func (command RepairCommand) readFiles(files map[string]os.FileInfo, path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			log.Warnf("'%v': permission denied", path)
			return nil
		case os.IsNotExist(err):
			return nil
		default:
			return fmt.Errorf("'%v': could not stat file: %v", path, err)
		}
	}

	files[path] = stat

	if stat.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("'%v': could not open file: %v", path, err)
		}

		childFilenames, err := file.Readdirnames(0)
		file.Close()
		if err != nil {
			return fmt.Errorf("'%v': could not read directory: %v", file.Name(), err)
		}

		for _, childFilename := range childFilenames {
			childPath := filepath.Join(path, childFilename)
			command.readFiles(files, childPath)
		}
	}

	return nil
}

func (command RepairCommand) addFile(store *storage.Storage, path string) (*database.File, error) {
	stat, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return nil, fmt.Errorf("'%v': permisison denied", path)
		case os.IsNotExist(err):
			return nil, fmt.Errorf("'%v': no such file", path)
		default:
			return nil, fmt.Errorf("'%v': could not stat file: %v", path, err)
		}
	}

	modTime := stat.ModTime().UTC()
	size := stat.Size()

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not create fingerprint: %v", path, err)
	}

	if !stat.IsDir() {
		duplicateCount, err := store.FileCountByFingerprint(fingerprint)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve count of files for fingerprint '%v': %v", path, fingerprint, err)
		}

		if duplicateCount > 0 {
			log.Info("'" + _path.Rel(path) + "' is a duplicate file.")
		}
	}

	file, err := store.AddFile(path, fingerprint, modTime, size)
	if err != nil {
		return nil, fmt.Errorf("'%v': could not add file: %v", path, err)
	}

	if stat.IsDir() {
		fsFile, err := os.Open(file.Path())
		if err != nil {
			return nil, fmt.Errorf("'%v': could not open file: %v", path, err)
		}
		defer fsFile.Close()

		dirFilenames, err := fsFile.Readdirnames(0)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not read dirctory listing: %v", path, err)
		}

		for _, dirFilename := range dirFilenames {
			dirFilePath := filepath.Join(path, dirFilename)
			if _, err = command.addFile(store, dirFilePath); err != nil {
				return nil, fmt.Errorf("'%v': could not add file: %v", dirFilePath, err)
			}
		}
	}

	return file, nil
}

func toMap(files database.Files) map[string]*database.File {
	fileByPath := make(map[string]*database.File)
	for _, file := range files {
		fileByPath[file.Path()] = file
	}
	return fileByPath
}
