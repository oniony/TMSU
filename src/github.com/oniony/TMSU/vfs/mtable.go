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

// +build !windows

package vfs

import (
	"bufio"
	"fmt"
	"github.com/oniony/TMSU/common/path"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Mount struct {
	DatabasePath string
	MountPath    string
}

func GetMountTable() ([]Mount, error) {
	mountTable := make([]Mount, 0, 10)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("could not open system mount table")
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for line, err := reader.ReadString('\n'); err != io.EOF; line, err = reader.ReadString('\n') {
		if err != nil {
			return nil, err
		}

		parts := strings.Split(line, " ")

		if parts[0] != "pathfs.pathInode" {
			continue
		}

		mountpoint := path.UnescapeOctal(parts[1])

		databaseSymlink := filepath.Join(mountpoint, ".database")
		databasePath, err := os.Readlink(databaseSymlink)
		if err != nil {
			return nil, err
		}

		mountTable = append(mountTable, Mount{databasePath, mountpoint})
	}

	return mountTable, nil
}
