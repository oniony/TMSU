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
	"tmsu/common"
	"tmsu/database"
)

type StatusCommand struct{}

func (StatusCommand) Name() string {
	return "status"
}

func (StatusCommand) Synopsis() string {
	return "List the status of database entries"
}

func (StatusCommand) Description() string {
	return `tmsu status [FILE]...

Shows the status of tagged entries in the database.

Where FILEs are given, only matching files are shown.

The status codes in the listing have the following meanings:

  T - Tagged
  M - Modified
  ! - Missing
  ? - Untagged

The 'repair' command can be used to fix problems caused by files that have been
modified or moved on disk.

  --disturbed    Show also files that have not been moved nor modified`
}

type StatusReport struct {
	Tagged   []string
	Modified []string
	Missing  []string
	Untagged []string
}

func NewReport() *StatusReport {
	return &StatusReport{make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10)}
}

func (command StatusCommand) Exec(args []string) error {
    all := false
    if len(args) > 0 && args[0] == "--all" {
        all = true
        args = args[1:]
    }

    report := NewReport()

    err := command.status(args, report)
    if err != nil {
        return err
    }

    if all {
        for _, path := range report.Tagged {
            fmt.Println("T", path)
        }
    }

	for _, path := range report.Modified {
        fmt.Println("M", path)
	}

	for _, path := range report.Missing {
        fmt.Println("!", path)
	}

	for _, path := range report.Untagged {
	    fmt.Println("?", path)
    }

	return nil
}

func (command StatusCommand) status(paths []string, report *StatusReport) (error) {
    if len(paths) == 0 {
        err := command.statusAll(report)
        if err != nil {
            return err
        }
    } else {
        for _, path := range paths {
            path = filepath.Clean(path)

            err := command.statusPath(path, report)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (command StatusCommand) statusPath(path string, report *StatusReport) (error) {
	db, err := database.OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

    entry, err := db.FileByPath(path)
    if err != nil {
        return err
    }
    if entry == nil {
        report.Untagged = append(report.Untagged, path)
    } else {
        command.fileSystemStatus(entry, report)
    }

    childEntries, err := db.FilesByDirectory(path)
    if err != nil {
        return err
    }

    for _, childEntry := range childEntries {
        command.fileSystemStatus(childEntry, report)
    }

    return nil
}

func (command StatusCommand) statusAll(report *StatusReport) (error) {
	db, err := database.OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

    entries, err := db.Files()
    if err != nil {
        return err
    }

	for _, entry := range entries {
	    command.fileSystemStatus(entry, report)
    }

    return nil
}

func (command StatusCommand) fileSystemStatus(entry *database.File, report *StatusReport) {
    fingerprint, err := common.Fingerprint(entry.Path())
    if err != nil {
        switch {
        case os.IsPermission(err):
            common.Warnf("'%v': Permission denied", entry.Path())
        case os.IsNotExist(err):
            report.Missing = append(report.Missing, entry.Path())
        default:
            common.Warnf("'%v': %v", entry.Path(), err)
        }
    } else {
        if entry.Fingerprint != fingerprint {
            report.Modified = append(report.Modified, entry.Path())
        } else {
            report.Tagged = append(report.Tagged, entry.Path())
        }
    }
}

func contains(strings []string, str string) bool {
    for _, item := range strings {
        if item == str {
            return true
        }
    }

    return false
}
