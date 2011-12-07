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
package main
*/

package main

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

const HALF_SECOND = 500000000

type MountCommand struct{}

func (this MountCommand) Name() string {
	return "mount"
}

func (this MountCommand) Summary() string {
	return "mounts the virtual file-system"
}

func (this MountCommand) Help() string {
	return `  tmsu mount MOUNTPOINT

Mounts the virtual file-system (VFS) at the mountpoint directory specified.
The default database at '$HOME/.tmsu/db' will be mounted unless overridden with the 'TMSU_DB' environment variable.`
}

func (this MountCommand) Exec(args []string) error {
	if len(args) < 1 { errors.New("No mountpoint specified.") }
	if len(args) > 1 { errors.New("Extraneous arguments.") }

    path := args[0]

    fileInfo, error := os.Stat(path)
    if error != nil { return error }
    if fileInfo == nil { return errors.New("Mount point '" + path + "' does not exist.") }
    if !fileInfo.IsDir() { return errors.New("Mount point '" + path + "' is not a directory.") }
    //TODO check permissions on mount path

	mountPath := args[0]
	command := exec.Command(os.Args[0], "vfs", databasePath(), mountPath)

	errorPipe, error := command.StderrPipe()
	if error != nil { return error }

	error = command.Start()
	if error != nil { return error }

    time.Sleep(HALF_SECOND)

    waitMessage, error := command.Process.Wait(os.WNOHANG)
    if error != nil { return error }

    if waitMessage.WaitStatus.Exited() {
        if waitMessage.WaitStatus.ExitStatus() != 0 {
            buffer := make([]byte, 1024)
            count, error := errorPipe.Read(buffer)
            if error != nil { return error }

            return errors.New("Could not mount VFS: " + string(buffer[0:count]))
        }
    }

	return nil
}
