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

package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"tmsu/common/proc"
)

type Mount struct {
	DatabasePath string
	MountPath    string
}

func GetMountTable() ([]Mount, error) {
	pids, err := proc.GetProcessIds()
	if err != nil {
		return nil, fmt.Errorf("could not get process IDs: %v", err)
	}

	mountTable := make([]Mount, 0, 10)

	for _, pid := range pids {
		process, err := proc.GetProcess(pid)
		if err != nil {
			if !os.IsPermission(err) {
				return nil, err
			}
		} else {
			if len(process.CommandLine) >= 4 && process.CommandLine[0] == "tmsu" && process.CommandLine[1] == "vfs" {
				databasePath := filepath.Join(process.WorkingDirectory, process.CommandLine[2])
				mountPath := process.CommandLine[3]
				if mountPath[0] != '/' {
					mountPath = filepath.Join(process.WorkingDirectory, mountPath)
				}
				mountTable = append(mountTable, Mount{databasePath, mountPath})
			}
		}
	}

	return mountTable, nil
}
