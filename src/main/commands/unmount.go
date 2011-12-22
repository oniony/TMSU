/*
Copyright 2011 Paul Ruane.

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

package main

import (
	"errors"
	"os"
	"os/exec"
)

type UnmountCommand struct{}

func (this UnmountCommand) Name() string {
	return "unmount"
}

func (this UnmountCommand) Summary() string {
	return "unmounts the virtual file-system"
}

func (this UnmountCommand) Help() string {
	return `  tags unount MOUNTPOINT

Unmounts a previously mounted virtual file-system at the mountpoint specified.`
}

func (this UnmountCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("Path to unmount not speciified.")
	}

	path := args[0]

	fusermountPath, error := exec.LookPath("fusermount")
	if error != nil { return error }

	process, error := os.StartProcess(fusermountPath, []string{fusermountPath, "-u", path}, &os.ProcAttr{})
	if error != nil { return error }

	message, error := os.Wait(process.Pid, 0)
	if error != nil { return error }
	if message.ExitStatus() != 0 {
		return errors.New("Could not unmount.")
	}

	return nil
}
