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

func (UnmountCommand) Name() string {
	return "unmount"
}

func (UnmountCommand) Synopsis() string {
	return "unmount the virtual file-system"
}

func (UnmountCommand) Description() string {
	return `tags unmount MOUNTPOINT

Unmounts the virtual file-system at MOUNTPOINT.`
}

func (UnmountCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("Path to unmount not speciified.")
	}

	path := args[0]

	fusermountPath, err := exec.LookPath("fusermount")
	if err != nil {
		return err
	}

	process, err := os.StartProcess(fusermountPath, []string{fusermountPath, "-u", path}, &os.ProcAttr{})
	if err != nil {
		return err
	}

	message, err := os.Wait(process.Pid, 0)
	if err != nil {
		return err
	}
	if message.ExitStatus() != 0 {
		return errors.New("Could not unmount.")
	}

	return nil
}
