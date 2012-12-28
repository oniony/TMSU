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

package proc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

type Process struct {
	Pid              int
	CommandLine      []string
	WorkingDirectory string
}

func GetProcessIds() ([]int, error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}

	dirNames, err := procDir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	procDir.Close()

	re, err := regexp.Compile("[0-9]+")
	if err != nil {
		return nil, err
	}

	pids := make([]int, 0, len(dirNames))

	for _, dirName := range dirNames {
		if re.MatchString(dirName) {
			pid, err := strconv.Atoi(dirName)
			if err != nil {
				return nil, err
			}

			pids = append(pids, pid)
		}
	}

	return pids, nil
}

func GetProcess(pid int) (*Process, error) {
	processDir := fmt.Sprintf("/proc/%v", pid)

	workingDirectory, err := os.Readlink(processDir + "/cwd")
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(processDir + "/cmdline")
	if err != nil {
		return nil, err
	}

	tokens := bytes.Split(data, []byte{0})
	commandLine := make([]string, len(tokens))
	for index, token := range tokens {
		commandLine[index] = string(token)
	}

	return &Process{pid, commandLine, workingDirectory}, nil
}
