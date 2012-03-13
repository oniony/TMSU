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
	"tmsu/fingerprint"
)

type StatusCommand struct{}

func (StatusCommand) Name() string {
	return "status"
}

func (StatusCommand) Synopsis() string {
	return "List the file tagging status"
}

func (StatusCommand) Description() string {
	return `tmsu status [PATH]...

Shows the status of entries in the database and file system.

Where no PATHs are given, the status of all the entries in the database are
listed. (Untagged files and directories are not identified in this case.)

Where PATHs are given then only the database entries corresponding to these
paths are shown, along with any untagged files on the filesystem at these paths.

The status codes in the listing have the following meanings:

  T - Tagged
  M - Modified
  ! - Missingd
  ? - Untagged
  + - Nested

The nested status is shown only for directories and indicates that the
directory is not tagged itself but that some of the files or directories within
it are.

Note: The 'repair' command can be used to fix problems caused by files that have
been modified or moved on disk.`
}

type StatusReport struct {
	Tagged   []string
	Modified []string
	Missing  []string
	Untagged []string
	Nested   []string
}

func NewReport() *StatusReport {
	return &StatusReport{make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10), make([]string, 0, 10)}
}

func (command StatusCommand) Exec(args []string) error {
    report := NewReport()

    err := command.status(args, report)
    if err != nil {
        return err
    }

    for _, path := range report.Tagged {
        fmt.Println("T", path)
    }

	for _, path := range report.Modified {
        fmt.Println("M", path)
	}

    for _, path := range report.Nested {
        fmt.Println("+", path)
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
	    db, err := database.Open()
        if err != nil {
            return err
        }
        defer db.Close()

        entries, err := db.Files()
        if err != nil {
            return err
        }

        for _, entry := range entries {
            paths = append(paths, entry.Path())
        }
    }

    for _, path := range paths {
        absPath, err := filepath.Abs(path)
        if err != nil {
            return err
        }

        status, err := command.statusPath(absPath)
        if err != nil {
            return err
        }

        relPath := common.MakeRelative(path)

        switch status {
        case UNTAGGED:
            report.Untagged = append(report.Untagged, relPath)
        case TAGGED:
            report.Tagged = append(report.Tagged, relPath)
        case MODIFIED:
            report.Modified = append(report.Modified, relPath)
        case MISSING:
            report.Missing = append(report.Missing, relPath)
        case NESTED:
            report.Nested = append(report.Nested, relPath)
        default:
            panic("Unsupported status " + string(status))
        }
    }

    return nil
}

func (command StatusCommand) statusPath(path string) (Status, error) {
	db, err := database.Open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

    entry, err := db.FileByPath(path)
    if err != nil {
        return 0, err
    }
    if entry != nil {
        // path is tagged

        _, err := os.Stat(path)
        if err != nil {
            switch {
            case os.IsNotExist(err):
                return MISSING, nil
            default:
                return 0, err
            }
        }

        fingerprint, err := fingerprint.Create(path)
        if err != nil {
            return 0, err
        }

        if entry.Fingerprint == fingerprint {
            return TAGGED, nil
        } else {
            return MODIFIED, nil
        }
    } else {
        // path is not tagged

        if common.IsDir(path) {
            dir, err := os.Open(path)
            if err != nil {
                return 0, err
            }

            entries, err := dir.Readdir(0)
            for _, entry := range entries {
                entryPath := filepath.Join(path, entry.Name())
                status, err := command.statusPath(entryPath)
                if err != nil {
                    return 0, err
                }

                switch status {
                case TAGGED, MODIFIED, NESTED:
                    return NESTED, err
                }
            }

            return UNTAGGED, err
        } else {
            return UNTAGGED, err
        }
    }

    return 0, nil
}

type Status int

const (
    UNTAGGED Status = iota
    TAGGED
    MODIFIED
    NESTED
    MISSING
)
