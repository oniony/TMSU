// Copyright 2011-2017 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"github.com/oniony/TMSU/common/log"
	_path "github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//TODO should return warnings for permission errors

var StatusCommand = Command{
	Name:     "status",
	Synopsis: "List the file tagging status",
	Usages:   []string{"tmsu status [PATH]..."},
	Description: `Shows the status of PATHs.

Where PATHs are not specified the status of the database is shown.

  T - Tagged
  M - Modified
  ! - Missing
  U - Untagged

Status codes of T, M and ! mean that the file has been tagged (and thus is in the TMSU database). Modified files are those with a different modification time or size to that in the database. Missing files are those in the database but that no longer exist in the file-system.

Note: The 'repair' subcommand can be used to fix problems caused by files that have been modified or moved on disk.`,
	Examples: []string{"$ tmsu status",
		"$ tmsu status .",
		"$ tmsu status --directory *"},
	Options: Options{Option{"--directory", "-d", "do not examine directory contents (non-recursive)", false, ""},
		Option{"--no-dereference", "-P", "do not follow symbolic links", false, ""}},
	Exec: statusExec,
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

// unexported

func statusExec(options Options, args []string, databasePath string) (error, warnings) {
	dirOnly := options.HasOption("--directory")
	followSymlinks := !options.HasOption("--no-dereference")

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	var report *StatusReport

	if len(args) == 0 {
		report, err = statusDatabase(store, tx, dirOnly, followSymlinks)
		if err != nil {
			return err, nil
		}
	} else {
		report, err = statusPaths(store, tx, args, dirOnly, followSymlinks)
		if err != nil {
			return err, nil
		}
	}

	printReport(report)

	return nil, nil
}

func statusDatabase(store *storage.Storage, tx *storage.Tx, dirOnly, followSymlinks bool) (*StatusReport, error) {
	report := NewReport()

	log.Info(2, "retrieving all files from database.")

	files, err := store.Files(tx, "name")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve files: %v", err)
	}

	err = statusCheckFiles(files, report)
	if err != nil {
		return nil, err
	}

	tree := _path.NewTree()
	for _, file := range files {
		tree.Add(file.Path(), file.IsDir)
	}

	topLevelPaths := tree.TopLevel().Paths()
	if err != nil {
		return nil, err
	}

	for _, path := range topLevelPaths {
		if err = findNewFiles(path, report, dirOnly, followSymlinks); err != nil {
			return nil, err
		}
	}

	return report, nil
}

func statusPaths(store *storage.Storage, tx *storage.Tx, paths []string, dirOnly, followSymlinks bool) (*StatusReport, error) {
	report := NewReport()

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("%v: could not get absolute path: %v", path, err)
		}

		log.Infof(2, "%v: resolving file", path)

		resolvedPath := absPath

		stat, err := os.Lstat(absPath)
		if err != nil {
			switch {
			case os.IsNotExist(err), os.IsPermission(err):
				stat = emptyStat{}
			default:
				return nil, fmt.Errorf("%v: could not stat path: %v", path, err)
			}
		} else if stat.Mode()&os.ModeSymlink != 0 {
			resolvedPath, err = _path.Dereference(absPath)
			if err != nil {
				return nil, fmt.Errorf("%v: could not dereference symbolic link: %v", path, err)
			}
		}

		log.Infof(2, "%v: checking file in database", path)

		file, err := store.FileByPath(tx, resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("%v: could not retrieve file: %v", path, err)
		}
		if file != nil {
			err = statusCheckFile(absPath, file, report)
			if err != nil {
				return nil, err
			}
		}

		if !dirOnly && (stat.Mode()&os.ModeSymlink == 0 || followSymlinks) {
			log.Infof(2, "%v: retrieving files from database.", path)

			files, err := store.FilesByDirectory(tx, resolvedPath)
			if err != nil {
				return nil, fmt.Errorf("%v: could not retrieve files for directory: %v", path, err)
			}

			err = statusCheckFiles(files, report)
			if err != nil {
				return nil, err
			}
		}

		err = findNewFiles(absPath, report, dirOnly, followSymlinks)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

func statusCheckFiles(files entities.Files, report *StatusReport) error {
	for _, file := range files {
		if err := statusCheckFile(file.Path(), file, report); err != nil {
			return err
		}
	}

	return nil
}

func statusCheckFile(absPath string, file *entities.File, report *StatusReport) error {
	log.Infof(2, "%v: checking file status.", absPath)

	stat, err := os.Stat(file.Path())
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Infof(2, "%v: file is missing.", absPath)

			report.AddRow(Row{absPath, MISSING})
			return nil
		case os.IsPermission(err):
			log.Warnf("%v: permission denied.", absPath)
		case strings.Contains(err.Error(), "not a directory"): //TODO improve
			report.AddRow(Row{file.Path(), MISSING})
			return nil
		default:
			return fmt.Errorf("%v: could not stat: %v", file.Path(), err)
		}
	} else {
		if stat.Size() != file.Size || !stat.ModTime().UTC().Equal(file.ModTime) {
			log.Infof(2, "%v: file is modified.", absPath)

			report.AddRow(Row{absPath, MODIFIED})
		} else {
			log.Infof(2, "%v: file is unchanged.", absPath)

			report.AddRow(Row{absPath, TAGGED})
		}
	}

	return nil
}

func findNewFiles(searchPath string, report *StatusReport, dirOnly, followSymlinks bool) error {
	log.Infof(2, "%v: finding new files.", searchPath)

	absPath, err := filepath.Abs(searchPath)
	if err != nil {
		return fmt.Errorf("%v: could not get absolute path: %v", searchPath, err)
	}

	if !report.ContainsRow(absPath) {
		report.AddRow(Row{absPath, UNTAGGED})
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

		sort.Strings(dirNames)

		for _, dirName := range dirNames {
			dirPath := filepath.Join(absPath, dirName)
			err = findNewFiles(dirPath, report, dirOnly, followSymlinks)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printReport(report *StatusReport) {
	printRows(report.Rows, TAGGED)
	printRows(report.Rows, MODIFIED)
	printRows(report.Rows, MISSING)
	printRows(report.Rows, UNTAGGED)
}

func printRows(rows []Row, status Status) {
	for _, row := range rows {
		if row.Status == status {
			printRow(row)
		}
	}
}

func printRow(row Row) {
	relPath := _path.Rel(row.Path)
	fmt.Printf("%v %v\n", string(row.Status), relPath)
}
