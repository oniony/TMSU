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
	"tmsu/common"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

//TODO need to remove implicitly tagged missing files right at the end otherwise moves not identified
//TODO need to look for new files right at the end otherwise moves are not identified

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
    4. Identifying missing files.
    5. Adding missing implicit taggings.

Where no PATHS are specified all tagged files and directories fingerprints in
the database are checked and their fingerprints updated where modifications are
found.

Modified files are identified by a change of the modification timestamp.

New files that lie under a tagged directory (and thus are implicitly tagged)
are added to the database.

Moved files are identified by file fingerprint and will only be found if they
have been moved under one of the specified PATHs. (As such, moved files cannot
be identified where no PATHs are specified.)

Files missing on disk are handled in two ways: those that are explicitly tagged
are reported as missing whilst those that are not are simply removed from the
database.`
}

func (RepairCommand) Options() cli.Options {
	return cli.Options{}
}

func (command RepairCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	if len(args) == 0 {
		return command.repairAll()
	}

	return command.repairPaths(args)
}

func (command RepairCommand) repairAll() error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.verbose {
		log.Info("retrieving all files from the database.")
	}

	files, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve files from storage: %v", err)
	}

	for _, file := range files {
		err := command.checkFile(file, store, make(map[int64][]string))
		if err != nil {
			return fmt.Errorf("could not check file #%v: %v", file.Id, err)
		}
	}

	return nil
}

func (command RepairCommand) repairPaths(paths []string) error {
	pathsBySize, err := command.buildFileSystemMap(paths)
	if err != nil {
		return err
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
		}

		if command.verbose {
			log.Infof("'%v': retrieving files from database.", path)
		}

		files, err := store.FilesByDirectory(absPath)
		if err != nil {
			return fmt.Errorf("'%v': could not retrieve files for directory from storage: %v", path, err)
		}

		for _, file := range files {
			err := command.checkFile(file, store, pathsBySize)
			if err != nil {
				return fmt.Errorf("could not check file #%v: %v", file.Id, err)
			}
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return fmt.Errorf("'%v': could not retrieve file: %v", path, err)
		}
		if file == nil {
			continue
		}

		err = command.checkFile(file, store, pathsBySize)
		if err != nil {
			return fmt.Errorf("'%v': could not check file: %v", path, err)
		}
	}

	return nil
}

func (command RepairCommand) checkFile(file *database.File, store *storage.Storage, pathsBySize map[int64][]string) error {
	if command.verbose {
		log.Infof("'%v': checking file status.", file.Path())
	}

	stat, err := os.Stat(file.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			err = command.processMissingfile(file, pathsBySize, store)
			if err != nil {
				return fmt.Errorf("'%v': could not process missing file: %v", file.Path(), err)
			}
		case os.IsPermission(err):
			log.Warnf("'%v': permission denied.", file.Path())
		default:
			log.Warnf("'%v': %v", file.Path(), err)
		}

		return nil
	}
	modTime := stat.ModTime().UTC()
	size := stat.Size()

	if modTime.Unix() != file.ModTimestamp.Unix() || size != file.Size {
		if command.verbose {
			log.Infof("'%v': updating file in database.", file.Path())
		}

		fingerprint, err := fingerprint.Create(file.Path())
		if err != nil {
			return err
		}

		_, err = store.UpdateFile(file.Id, file.Path(), fingerprint, modTime, size)
		if err != nil {
			return err
		}

		log.Printf("'%v': modified.", file.Path())
	} else {
		if command.verbose {
			log.Infof("'%v': unchanged.", file.Path())
		}
	}

	if stat.IsDir() {
		tags, err := store.TagsByFileId(file.Id)
		if err != nil {
			return err
		}

		err = command.processDirectory(store, file, tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command RepairCommand) processDirectory(store *storage.Storage, file *database.File, tags database.Tags) error {
	dir, err := os.Open(file.Path())
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
		childPath := filepath.Join(file.Path(), filename)

		childFile, err := store.FileByPath(childPath)
		if err != nil {
			return err
		}
		if childFile == nil {
			childFile, err = command.addFile(store, childPath)
			if err != nil {
				return err
			}

			log.Printf("'%v': added.", childPath)
		}

		for _, tag := range tags {
			if command.verbose {
				log.Infof("'%v': ensuring implicit tagging '%v' exists.", childPath, tag.Name)
			}

			_, err := store.AddImplicitFileTag(childFile.Id, tag.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (command RepairCommand) processMissingfile(file *database.File, pathsBySize map[int64][]string, store *storage.Storage) error {
	if file.Fingerprint == "" {
		if command.verbose {
			log.Infof("'%v': not searching for new location (no fingerprint).", file.Path())
		}

		return nil
	}

	if command.verbose {
		log.Infof("'%v': searching for new location.", file.Path())
	}

	paths, sizeFound := pathsBySize[file.Size]
	if sizeFound {
		for _, path := range paths {
			fingerprint, err := fingerprint.Create(path)
			if err != nil {
				return err
			}

			if fingerprint == file.Fingerprint {
				if command.verbose {
					log.Infof("'%v': file with same fingerprint found at '%v'.", file.Path(), path)
				}

				stat, err := os.Stat(path)
				if err != nil {
					return err
				}

				_, err = store.UpdateFile(file.Id, path, file.Fingerprint, stat.ModTime().UTC(), stat.Size())
				if err != nil {
					return err
				}

				log.Printf("'%v': moved to '%v'.", file.Path(), path)
				return nil
			}
		}
	}

	explicitFileTagCount, err := store.ExplicitFileTagCountByFileId(file.Id)
	if err != nil {
		return fmt.Errorf("'%v': could not get explicit file-tag count: %v", file.Path(), err)
	}

	if explicitFileTagCount == 0 {
		err := store.RemoveImplicitFileTagsByFileId(file.Id)
		if err != nil {
			return fmt.Errorf("'%v': could not remove implicit file tags: %v", file.Path(), err)
		}

		err = store.RemoveFile(file.Id)
		if err != nil {
			return fmt.Errorf("'%v': could not remove file: %v", file.Path(), err)
		}

		log.Printf("'%v': removed.", file.Path())
	} else {
		log.Printf("'%v': missing.", file.Path())
	}

	return nil
}

func (command RepairCommand) buildFileSystemMap(paths []string) (map[int64][]string, error) {
	if command.verbose {
		log.Infof("building map of files by size.")
	}

	pathsBySize := make(map[int64][]string)

	for _, path := range paths {
		err := command.buildFileSystemMapRecursive(path, pathsBySize)
		if err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("'%v': permission denied.")
				continue
			}

			return nil, fmt.Errorf("could not build filesystem map: %v", err)
		}
	}

	return pathsBySize, nil
}

func (command RepairCommand) buildFileSystemMapRecursive(path string, pathsBySize map[int64][]string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("'%v': could not stat file: %v", path, err)
	}

	if stat.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("'%v': could not open file: %v", path, err)
		}

		dirFiles, err := file.Readdir(0)
		if err != nil {
			return fmt.Errorf("'%v': could not read directory listing: %v", path, err)
		}

		file.Close()

		for _, dirfile := range dirFiles {
			dirFilePath := filepath.Join(path, dirfile.Name())
			err := command.buildFileSystemMapRecursive(dirFilePath, pathsBySize)
			if err != nil {
				return fmt.Errorf("'%v': could not build filesystem map: %v", dirFilePath, err)
			}
		}
	} else {
		if stat.Size() > 0 {
			paths, found := pathsBySize[stat.Size()]
			if !found {
				paths = make([]string, 0, 10)
			}
			paths = append(paths, path)
			pathsBySize[stat.Size()] = paths
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
			log.Info("'" + common.RelPath(path) + "' is a duplicate file.")
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
