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

package vfs

import (
    "bytes"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"
    "tmsu/common"
)

type Mount struct {
    DatabasePath string
    MountPath    string
}

func GetMountTable() ([]Mount, error) {
    //TODO change this to examine /proc directly

    outputBytes, err := exec.Command("pgrep", "-f", "tmsu vfs").Output()
    if err != nil {
        return []Mount{}, nil //TODO currently assumes no matches
    }

    pids := strings.Split(strings.Trim(string(outputBytes), "\n"), "\n")

    mountTable := make([]Mount, len(pids))

    for index, pid := range pids {
        metaPath := "/proc/" + pid

        workingDirectory, err := os.Readlink(metaPath + "/cwd")
        if err != nil {
            return nil, err
        }

        data, err := ioutil.ReadFile(metaPath + "/cmdline")
        if err != nil {
            return nil, err
        }

        tokens := bytes.Split(data, []byte{0})

        databasePath := common.Join(workingDirectory, string(tokens[2]))
        mountPath := common.Join(workingDirectory, string(tokens[3]))
        mountTable[index] = Mount{databasePath, mountPath}
    }

    return mountTable, nil
}
