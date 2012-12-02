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
	"strconv"
	"tmsu/cli"
	"tmsu/common"
	"tmsu/database"
)

type StatusCommand struct{}

func (StatusCommand) Name() string {
	return "status"
}

func (StatusCommand) Synopsis() string {
	return "List the file tagging status"
}

func (StatusCommand) Description() string {
	return `tmsu status [--directory] [PATH]...

Shows the status of PATHs.

Where PATHs are not specified, the statuses of the contents of the working
directory are shown.

  --directory    list directory entries instead of contents

Status codes are shown in the first column:

  T - Tagged
  M - Modified
  ! - Missing
  U - Untagged

If the status code is followed by a plus (+) this indicates that it is a
directory containing one or more tagged items.

Note: The 'repair' command can be used to fix problems caused by files that have
been modified or moved on disk.`
}

type StatusReport struct {
	Rows []Row
}

type Row struct {
	Path   string
	Status Status
	Nested bool
}

func NewReport() *StatusReport {
	return &StatusReport{make([]Row, 0, 10)}
}

func (StatusCommand) Options() []cli.Option {
	return []cli.Option{}
}

func (command StatusCommand) Exec(args []string) error {
	showDirectory := false

	if len(args) > 0 && args[0] == "--directory" {
		showDirectory = true
		args = args[1:]
	}

	report := NewReport()

	err := command.status(args, report, showDirectory)
	if err != nil {
		return err
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

func (command StatusCommand) status(paths []string, report *StatusReport, showDirectory bool) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	db, err := database.Open()
	if err != nil {
		return err
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		status, nested, err := command.getStatus(absPath, db)
		if err != nil {
			return err
		}

		report.Rows = append(report.Rows, Row{path, status, nested})

		if !showDirectory && isDir(absPath) {
			dir, err := os.Open(absPath)
			if err != nil {
				return err
			}
			defer dir.Close()

			entryNames, err := dir.Readdirnames(0)
			for _, entryName := range entryNames {
				entryAbsPath := filepath.Join(absPath, entryName)
				entryRelPath := common.RelPath(entryAbsPath)

				status, nested, err := command.getStatus(entryAbsPath, db)
				if err != nil {
					return err
				}

				report.Rows = append(report.Rows, Row{entryRelPath, status, nested})
			}

			files, err := db.FilesByDirectory(absPath)
			for _, file := range files {
				fileAbsPath := file.Path()
				fileRelPath := common.RelPath(fileAbsPath)

				status, nested, err := command.getStatus(fileAbsPath, db)
				if err != nil {
					return err
				}

				if status == MISSING {
					report.Rows = append(report.Rows, Row{fileRelPath, status, nested})
				}
			}
		}
	}

	return nil
}

func (command StatusCommand) getStatus(path string, db *database.Database) (Status, bool, error) {
	entry, err := db.FileByPath(path)
	if err != nil {
		return 0, false, err
	}

	var status Status
	if entry != nil {
		info, err := os.Stat(path)
		if err != nil {
			return 0, false, nil
		}

		if entry.ModTimestamp.Unix() == info.ModTime().Unix() {
			status = TAGGED
		} else {
			status = MODIFIED
		}
	} else {
		status = UNTAGGED
	}

	nested, err := command.isNested(path, db)
	if err != nil {
		return 0, false, err
	}

	return status, nested, nil
}

func (StatusCommand) printRow(row Row) {
	statusCode := getStatusCode(row.Status)
	nestedCode := getNestedCode(row.Nested)
	path := row.Path

	fmt.Printf("%v%v %v\n", statusCode, nestedCode, path)
}

func (command StatusCommand) isNested(path string, db *database.Database) (bool, error) {
	if !common.IsDir(path) {
		return false, nil
	}

	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}

	entries, err := dir.Readdir(0)
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		status, nested, err := command.getStatus(entryPath, db)
		if err != nil {
			return false, err
		}

		switch status {
		case TAGGED, MODIFIED, MISSING:
			return true, nil
		}

		if nested {
			return true, nil
		}
	}

	return false, nil
}

func getStatusCode(status Status) string {
	switch status {
	case TAGGED:
		return "T"
	case MODIFIED:
		return "M"
	case MISSING:
		return "!"
	case UNTAGGED:
		return "U"
	}

	panic("Unsupported status '" + strconv.Itoa(int(status)) + "'.")
}

func getNestedCode(nested bool) string {
	if nested {
		return "+"
	}
	return " "
}

//TODO this needs to look in the database rather than the file-system
//     otherwise it will incorrectly report for directories with tagged
//     contents that have been replaced with identically named file
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

type Status int

const (
	UNTAGGED Status = iota
	TAGGED
	MODIFIED
	MISSING
)
