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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tmsu/entities"
	"tmsu/log"
	"tmsu/path"
	"tmsu/storage"
)

var StatusCommand = Command{
	Name:     "status",
	Synopsis: "List the file tagging status",
	Description: `tmsu status [PATH]...

Shows the status of PATHs.

Where PATHs are not specified the status of the database is shown.

  T - Tagged
  M - Modified
  ! - Missing
  U - Untagged

Status codes of T, M and ! mean that the file has been tagged (and thus is in
the TMSU database). Modified files are those with a different modification time
or size to that in the database. Missing files are those in the database but
that no longer exist in the file-system.

Note: The 'repair' command can be used to fix problems caused by files that have
been modified or moved on disk.`,
	Options: Options{Option{"--directory", "-d", "list directory entries only: do not list contents", false, ""}},
	Exec:    statusExec,
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

func statusExec(options Options, args []string) error {
	dirOnly := options.HasOption("--directory")

	var report *StatusReport
	var err error

	if len(args) == 0 {
		report, err = statusDatabase(dirOnly)
		if err != nil {
			return err
		}
	} else {
		report, err = statusPaths(args, dirOnly)
		if err != nil {
			return err
		}
	}

	printReport(report)

	return nil
}

func statusDatabase(dirOnly bool) (*StatusReport, error) {
	report := NewReport()

	store, err := storage.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "retrieving all files from database.")

	files, err := store.Files()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve files: %v", err)
	}

	err = statusCheckFiles(files, report)
	if err != nil {
		return nil, err
	}

	tree := path.NewTree()
	for _, file := range files {
		tree.Add(file.Path(), file.IsDir)
	}

	topLevelPaths := tree.TopLevel().Paths()
	if err != nil {
		return nil, err
	}

	for _, path := range topLevelPaths {
		if err = findNewFiles(path, report, dirOnly); err != nil {
			return nil, err
		}
	}

	return report, nil
}

func statusPaths(paths []string, dirOnly bool) (*StatusReport, error) {
	report := NewReport()

	store, err := storage.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		file, err := store.FileByPath(absPath)
		if err != nil {
			return nil, fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file != nil {
			err = statusCheckFile(file, report)
			if err != nil {
				return nil, err
			}
		}

		if !dirOnly {
			log.Infof(2, "%v: retrieving files from database.", path)

			files, err := store.FilesByDirectory(absPath)
			if err != nil {
				return nil, fmt.Errorf("%v: could not retrieve files for directory: %v", path, err)
			}

			err = statusCheckFiles(files, report)
			if err != nil {
				return nil, err
			}
		}

		err = findNewFiles(path, report, dirOnly)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

func statusCheckFiles(files entities.Files, report *StatusReport) error {
	for _, file := range files {
		err := statusCheckFile(file, report)
		if err != nil {
			return err
		}
	}

	return nil
}

func statusCheckFile(file *entities.File, report *StatusReport) error {
	relPath := path.Rel(file.Path())

	log.Infof(2, "%v: checking file status.", file.Path())

	stat, err := os.Stat(file.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Infof(2, "%v: file is missing.", file.Path())

			report.AddRow(Row{relPath, MISSING})
			return nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied.", file.Path())
		case strings.Contains(err.Error(), "not a directory"):
			report.AddRow(Row{relPath, MISSING})
			return nil
		default:
			return fmt.Errorf("%v: could not stat: %v", file.Path(), err)
		}
	} else {
		if stat.Size() != file.Size || stat.ModTime().UTC() != file.ModTime {
			log.Infof(2, "%v: file is modified.", file.Path())

			report.AddRow(Row{relPath, MODIFIED})
		} else {
			log.Infof(2, "%v: file is unchanged.", file.Path())

			report.AddRow(Row{relPath, TAGGED})
		}
	}

	return nil
}

func findNewFiles(searchPath string, report *StatusReport, dirOnly bool) error {
	log.Infof(2, "%v: finding new files.", searchPath)

	relPath := path.Rel(searchPath)

	if !report.ContainsRow(relPath) {
		report.AddRow(Row{relPath, UNTAGGED})
	}

	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", searchPath, err)
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied.", searchPath)
			return nil
		default:
			return fmt.Errorf("%v: could not stat: %v", searchPath, err)
		}
	}

	if !dirOnly && stat.IsDir() {
		dir, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("%v: could not open file: %v", searchPath, err)
		}

		dirNames, err := dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("%v: could not read directory listing: %v", searchPath, err)
		}

		for _, dirName := range dirNames {
			dirPath := filepath.Join(searchPath, dirName)
			err = findNewFiles(dirPath, report, dirOnly)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printReport(report *StatusReport) {
	for _, row := range report.Rows {
		if row.Status == TAGGED {
			printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == MODIFIED {
			printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == MISSING {
			printRow(row)
		}
	}

	for _, row := range report.Rows {
		if row.Status == UNTAGGED {
			printRow(row)
		}
	}
}

func printRow(row Row) {
	fmt.Printf("%v %v\n", string(row.Status), row.Path)
}
