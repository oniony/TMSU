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
	return `tmsu mount [OPTION] MOUNTPOINT

Mounts the currently selected database at MOUNTPOINT.

  --name NAME   Mounts the configured database NAME instead.`
}

func (command MountCommand) Exec(args []string) error {
    argCount := len(args)

    if argCount < 1 { return errors.New("Mountpoint must be specified.") }

    if argCount > 0 && args[0] == "--name" {
        if argCount < 3 { return errors.New("Database name and mountpoint must be specified.") }
        if argCount > 3 { return errors.New("Too many arguments.") }

        name := args[1]
        mountPath := args[2]

        err := command.mountNamed(name, mountPath)
        if err != nil { return errors.New("Could not mount database: " + err.Error()) }
    } else {
        if argCount < 1 { return errors.New("Mountpoint must be specified.") }
        if argCount > 1 { return errors.New("Too many arguments.") }

        mountPath := args[0]

        err := command.mountSelected(mountPath)
        if err != nil { return errors.New("Could not mount currently selected database: " + err.Error()) }
    }

    return nil
}

func (command MountCommand) mountNamed(name, mountPath string) error {
    config, err := GetDatabaseConfig(name)
    if err != nil { return err }
    if config == nil { return errors.New("No configured database called '" + name + "'.") }

    err = command.mountExplicit(config.DatabasePath, mountPath)
    if err != nil { return err }

    return nil
}

func (command MountCommand) mountSelected(mountPath string) error {
    config, err := GetSelectedDatabaseConfig()
    if err != nil { return err }
    if config == nil { return errors.New("Could not get selected database configuration.") }

    err = command.mountExplicit(config.DatabasePath, mountPath)
    if err != nil { return err }

    return nil
}

func (MountCommand) mountExplicit(databasePath string, mountPath string) error {
    fileInfo, err := os.Stat(mountPath)
    if err != nil { return err }
    if fileInfo == nil { return errors.New("Mount point '" + mountPath + "' does not exist.") }
    if !fileInfo.IsDir() { return errors.New("Mount point '" + mountPath + "' is not a directory.") }

    fileInfo, err = os.Stat(databasePath)
    if err != nil { return err }
    if fileInfo == nil { return errors.New("Database '" + databasePath + "' does not exist.") }

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
