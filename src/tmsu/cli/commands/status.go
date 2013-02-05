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
	"tmsu/log"
	"tmsu/path"
	"tmsu/storage"
	"tmsu/storage/database"
)

type StatusCommand struct {
	verbose bool
}

func (StatusCommand) Name() cli.CommandName {
	return "status"
}

func (StatusCommand) Synopsis() string {
	return "List the file tagging status"
}

func (StatusCommand) Description() string {
	return `tmsu status [PATH]...

Shows the status of PATHs.

Where PATHs are not specified the status of the database is shown.

  T - Tagged
  M - Modified
  ! - Missing
  U - Untagged

Status codes of T, M and ! mean that the file has been tagged (and thus is in
the TMSU database). Modified files are those with a different fingerprint to
that in the database. Missing files are those in the database but no longer on
in the file system.

Note: The 'repair' command can be used to fix problems caused by files that have
been modified or moved on disk.`
}

type Status byte

const (
	UNTAGGED Status = 'U'
	TAGGED   Status = 'T'
	MODIFIED Status = 'M'
	MISSING  Status = '!'
)

type StatusReport struct {
	Rows []Row
}

func (report *StatusReport) AddRow(row Row) {
	report.Rows = append(report.Rows, row)
}

func (report *StatusReport) ContainsRow(path string) bool {
	for _, row := range report.Rows {
		if row.Path == path {
			return true
		}
	}

	return false
}

type Row struct {
	Path   string
	Status Status
}

func NewReport() *StatusReport {
	return &StatusReport{make([]Row, 0, 10)}
}

func (StatusCommand) Options() cli.Options {
	return cli.Options{}
}

func (command StatusCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	var report *StatusReport
	var err error

	if len(args) == 0 {
		report, err = command.statusDatabase()
		if err != nil {
			return err
		}
	} else {
		report, err = command.statusPaths(args)
		if err != nil {
			return err
		}
	}

	for _, row := range report.Rows {
		if row.Status == TAGGED {
			command.printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == MODIFIED {
			command.printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == MISSING {
			command.printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == UNTAGGED {
			command.printRow(row)
		}
	}

	return nil
}

func (command StatusCommand) statusDatabase() (*StatusReport, error) {
	report := NewReport()

	store, err := storage.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if command.verbose {
		log.Info("retrieving all files from database.")
	}

	files, err := store.Files()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve files: %v", err)
	}

	err = command.checkFiles(files, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (command StatusCommand) statusPaths(paths []string) (*StatusReport, error) {
	report := NewReport()

	store, err := storage.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not get absolute path: %v", path, err)
		}

		if path != "." {
			file, err := store.FileByPath(absPath)
			if err != nil {
				return nil, fmt.Errorf("'%v': could not retrieve file: %v", path, err)
			}
			if file != nil {
				err = command.checkFile(file, report)
				if err != nil {
					return nil, err
				}
			}
		}

		if command.verbose {
			log.Infof("'%v': retrieving files from database.", path)
		}

		files, err := store.FilesByDirectory(absPath)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not retrieve files for directory: %v", path, err)
		}

		err = command.checkFiles(files, report)
		if err != nil {
			return nil, err
		}

		err = command.findNewFiles(path, report)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

func (command *StatusCommand) checkFiles(files database.Files, report *StatusReport) error {
	for _, file := range files {
		err := command.checkFile(file, report)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command *StatusCommand) checkFile(file *database.File, report *StatusReport) error {
	relPath := path.Rel(file.Path())

	if command.verbose {
		log.Infof("'%v': checking file status.", file.Path())
	}

	stat, err := os.Stat(file.Path())
	if err != nil {
		pathError := err.(*os.PathError)

		switch {
		case os.IsNotExist(pathError.Err):
			if command.verbose {
				log.Infof("'%v': file is missing.", file.Path())
			}

			report.AddRow(Row{relPath, MISSING})
			return nil
		case os.IsPermission(pathError.Err):
			log.Warnf("%v: Permission denied.", file.Path())
		default:
			return fmt.Errorf("'%v': could not stat: %v", file.Path(), err)
		}
	} else {
		if stat.Size() != file.Size || stat.ModTime().UTC() != file.ModTime {
			if command.verbose {
				log.Infof("'%v': file is modified.", file.Path())
			}

			report.AddRow(Row{relPath, MODIFIED})
		} else {
			if command.verbose {
				log.Infof("'%v': file is unchanged.", file.Path())
			}

			report.AddRow(Row{relPath, TAGGED})
		}
	}

	return nil
}

func (command *StatusCommand) findNewFiles(path string, report *StatusReport) error {
	if command.verbose {
		log.Infof("'%v': finding new files.", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("'%v': could not get absolute path: %v", path, err)
	}

	if !report.ContainsRow(path) {
		report.AddRow(Row{path, UNTAGGED})
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		pathError := err.(*os.PathError)

		switch {
		case os.IsNotExist(pathError.Err):
			return nil
		case os.IsPermission(pathError.Err):
			log.Warnf("%v: Permission denied.", path)
			return nil
		default:
			return fmt.Errorf("'%v': could not stat: %v", path, err)
		}
	}

	if stat.IsDir() {
		dir, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("'%v': could not open file: %v", path, err)
		}

		dirNames, err := dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("'%v': could not read directory listing: %v", path, err)
		}

		for _, dirName := range dirNames {
			dirPath := filepath.Join(path, dirName)
			err = command.findNewFiles(dirPath, report)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (command *StatusCommand) printRow(row Row) {
	log.Printf("%v %v", string(row.Status), row.Path)
}
