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
	return `  tmsu mount NAME
tmsu mount FILE MOUNTPOINT

In its first form, mounts the database NAME using the database file path and
mountpoint path configured in the application configuration file.

In its second form, mounts the database FILE at the mountpoint MOUNTPOINT. (In
this form the database need not be present in the configuration file.)`
}

func (command MountCommand) Exec(args []string) error {
	if len(args) < 1 { return errors.New("Not enough arguments.") }
	if len(args) > 2 { return errors.New("Too many arguments.") }

    if len(args) == 1 { return command.mountPreconfigured(args[0]) }

    return command.mount(args[0], args[1])
}

func (MountCommand) mountPreconfigured(name string) error {
    //TODO read configuration file
    //TODO find database with same name
    //TODO     error if not found
    //TODO chain to mount(string,string)
    return nil
}

func (MountCommand) mount(databasePath string, mountPath string) error {
    fileInfo, err := os.Stat(mountPath)
    if err != nil { return err }
    if fileInfo == nil { return errors.New("Mount point '" + mountPath + "' does not exist.") }
    if !fileInfo.IsDir() { return errors.New("Mount point '" + mountPath + "' is not a directory.") }

	command := exec.Command(os.Args[0], "vfs", databasePath, mountPath)

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
