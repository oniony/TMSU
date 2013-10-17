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
	"tmsu/storage/entities"
)

type RepairCommand struct {
	verbose bool
	pretend bool
	force   bool
}

func (RepairCommand) Name() cli.CommandName {
	return "repair"
}

func (RepairCommand) Synopsis() string {
	return "Repair the database"
}

func (RepairCommand) Description() string {
	return `tmsu [OPTION]... repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

Where no PATHS are specified all files in the database are checked.

                                          Reported Repaired
    Modified files                          yes      yes
    Moved files                             yes      yes
    Missing files                           yes      no
    Untagged files                          yes      no

Modified files are identified by a change to the file's modification time or
file size. These files are repaired by updating the modification time, size and
fingerprint in the database.

Moved files will only be repaired if a file with the same fingerprint can be
found under PATHs: this means files that are simultaneously moved and modified
will not be identified. Where no PATHs are specified, moved files will only be
identified if moved to a tagged directory.

Missing files are reported but are not, by default, removed from the database
as this would destroy the tagging information associated with them. If you do
wish to clear missing files from the database and destroying the associated
tagging information then use the --force option.

Untagged files are reported but not added to the database.`
}

func (RepairCommand) Options() cli.Options {
	return cli.Options{{"--pretend", "-p", "do not make any changes", false, ""},
		{"--force", "-f", "remove missing files from the database", false, ""}}
}

func (command RepairCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.pretend = options.HasOption("--pretend")
	command.force = options.HasOption("--force")

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

//- unexported

func (command RepairCommand) repairDatabase(store *storage.Storage) error {
	if command.verbose {
		log.Info("retrieving all files from the database.")
	}

	files, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve paths from storage: %v", err)
	}

	paths := make([]string, len(files))
	for index := 0; index < len(files); index++ {
		paths[index] = files[index].Path()
	}

	err = command.checkFiles(store, paths)
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
			return fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		absPaths[index] = absPath
	}

	if command.verbose {
		log.Infof("identifying top-level paths.")
	}

	err := command.checkFiles(store, absPaths)
	if err != nil {
		return err
	}

	return nil
}

func (command RepairCommand) checkFiles(store *storage.Storage, paths []string) error {
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

	_, untagged, modified, missing := command.determineStatuses(fsPaths, dbPaths)

	if err = command.repairModified(store, modified); err != nil {
		return err
	}

	if err = command.repairMoved(store, missing, untagged); err != nil {
		return err
	}

	if err = command.repairMissing(store, missing); err != nil {
		return err
	}

	for path, _ := range untagged {
		log.Infof("%v: untagged", path)
	}

	//TODO cleanup: any files that have no tags: remove
	//TODO cleanup: any tags that do not correspond to a file: remove

	return nil
}

type fileInfoMap map[string]os.FileInfo
type fileIdAndInfoMap map[string]struct {
	fileId uint
	stat   os.FileInfo
}
type databaseFileMap map[string]entities.File

func (command RepairCommand) determineStatuses(fsPaths fileInfoMap, dbPaths databaseFileMap) (tagged databaseFileMap, untagged fileInfoMap, modified fileIdAndInfoMap, missing databaseFileMap) {
	if command.verbose {
		log.Info("determining file statuses")
	}

	tagged = make(databaseFileMap, 100)
	untagged = make(fileInfoMap, 100)
	modified = make(fileIdAndInfoMap, 100)
	missing = make(databaseFileMap, 100)

	for path, stat := range fsPaths {
		if dbFile, isTagged := dbPaths[path]; isTagged {
			if dbFile.ModTime == stat.ModTime().UTC() && dbFile.Size == stat.Size() {
				tagged[path] = dbFile
			} else {
				modified[path] = struct {
					fileId uint
					stat   os.FileInfo
				}{dbFile.Id, stat}
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

	return tagged, untagged, modified, missing
}

func (command RepairCommand) repairModified(store *storage.Storage, modified fileIdAndInfoMap) error {
	if command.verbose {
		log.Info("repairing modified files")
	}

	for path, fileIdAndStat := range modified {
		fileId := fileIdAndStat.fileId
		stat := fileIdAndStat.stat

		log.Infof("%v: modified", path)

		fingerprint, err := fingerprint.Create(path)
		if err != nil {
			return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
		}

		if !command.pretend {
			_, err := store.UpdateFile(fileId, path, fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", path, err)
			}
		}

	}

	return nil
}

func (command RepairCommand) repairMoved(store *storage.Storage, missing databaseFileMap, untagged fileInfoMap) error {
	if command.verbose {
		log.Info("repairing moved files")
	}

	moved := make([]string, 0, 10)

	for path, dbFile := range missing {
		if command.verbose {
			log.Infof("%v: searching for new location", path)
		}

		for candidatePath, stat := range untagged {
			if stat.Size() == dbFile.Size {
				fingerprint, err := fingerprint.Create(candidatePath)
				if err != nil {
					return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
				}

				if fingerprint == dbFile.Fingerprint {
					log.Infof("%v: moved to %v", path, candidatePath)

					moved = append(moved, path)

					if !command.pretend {
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

func (command RepairCommand) repairMissing(store *storage.Storage, missing databaseFileMap) error {
	for path, dbFile := range missing {
		if command.force && !command.pretend {
			if err := store.RemoveFileTagsByFileId(dbFile.Id); err != nil {
				return fmt.Errorf("%v: could not delete file-tags: %v", path, err)
			}

			if err := store.RemoveFile(dbFile.Id); err != nil {
				return fmt.Errorf("%v: could not delete file: %v", path, err)
			}

			log.Infof("%v: removed", path)
		} else {
			log.Infof("%v: missing", path)
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
