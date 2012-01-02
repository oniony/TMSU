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
	"time"
)

type MountCommand struct{}

func (MountCommand) Name() string {
	return "mount"
}

func (MountCommand) Summary() string {
	return "mounts the virtual file-system"
}

func (MountCommand) Help() string {
	return `  tmsu mount MOUNTPOINT

Mounts the virtual file-system (VFS) at the mountpoint directory specified.
The default database at '$HOME/.tmsu/db' will be mounted unless overridden with the 'TMSU_DB' environment variable.`
}

func (MountCommand) Exec(args []string) error {
	if len(args) < 1 { return errors.New("No mountpoint specified.") }
	if len(args) > 1 { return errors.New("Extraneous arguments.") }

    path := args[0]

    fileInfo, err := os.Stat(path)
    if err != nil { return err }
    if fileInfo == nil { return errors.New("Mount point '" + path + "' does not exist.") }
    if !fileInfo.IsDir() { return errors.New("Mount point '" + path + "' is not a directory.") }

	mountPath := args[0]
	command := exec.Command(os.Args[0], "vfs", databasePath(), mountPath)

	errorPipe, err := command.StderrPipe()
	if err != nil { return err }

	err = command.Start()
	if err != nil { return err }

    const HALF_SECOND = 500000000
    time.Sleep(HALF_SECOND)

    waitMessage, err := command.Process.Wait(os.WNOHANG)
    if err != nil { return err }

    if waitMessage.WaitStatus.Exited() {
        if waitMessage.WaitStatus.ExitStatus() != 0 {
            buffer := make([]byte, 1024)
            count, err := errorPipe.Read(buffer)
            if err != nil { return err }

            return errors.New("Could not mount VFS: " + string(buffer[0:count]))
        }
    }

	return nil
}
