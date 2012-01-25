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

func (MountCommand) Synopsis() string {
	return "mount the virtual file-system"
}

func (MountCommand) Description() string {
	return `tmsu mount [FILE] MOUNTPOINT

Mounts a virtual file-system at the path MOUNTPOINT.

Where FILE is specified, the database at FILE is mounted.

If FILE is not specified but the TMSU_DB environment variable is defined then
the database at TMSU_DB is mounted.

Where neither FILE is specified nor TMSU_DB defined then the default database
is mounted.`
}

func (command MountCommand) Exec(args []string) error {
	argCount := len(args)

	switch argCount {
	case 0:
		return errors.New("Mountpoint must be specified.")
	case 1:
		mountPath := args[0]

		err := command.mountSelected(mountPath)
		if err != nil {
			return errors.New("Could not mount database: " + err.Error())
		}
	case 2:
		databasePath := args[0]
		mountPath := args[1]

		err := command.mountExplicit(databasePath, mountPath)
		if err != nil {
			return errors.New("Could not mount database: " + err.Error())
		}
	default:
		return errors.New("Too many arguments.")
	}

	return nil
}

func (command MountCommand) mountSelected(mountPath string) error {
	config, err := GetSelectedDatabaseConfig()
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("Could not get selected database configuration.")
	}

	err = command.mountExplicit(config.DatabasePath, mountPath)
	if err != nil {
		return err
	}

	return nil
}

func (MountCommand) mountExplicit(databasePath string, mountPath string) error {
	fileInfo, err := os.Stat(mountPath)
	if err != nil {
		return err
	}
	if fileInfo == nil {
		return errors.New("Mount point '" + mountPath + "' does not exist.")
	}
	if !fileInfo.IsDir() {
		return errors.New("Mount point '" + mountPath + "' is not a directory.")
	}

	fileInfo, err = os.Stat(databasePath)
	if err != nil {
		return err
	}
	if fileInfo == nil {
		return errors.New("Database '" + databasePath + "' does not exist.")
	}

	command := exec.Command(os.Args[0], "vfs", databasePath, mountPath)

	errorPipe, err := command.StderrPipe()
	if err != nil {
		return err
	}

	err = command.Start()
	if err != nil {
		return err
	}

	const HALF_SECOND = 500000000
	time.Sleep(HALF_SECOND)

	waitMessage, err := command.Process.Wait(os.WNOHANG)
	if err != nil {
		return err
	}

	if waitMessage.WaitStatus.Exited() {
		if waitMessage.WaitStatus.ExitStatus() != 0 {
			buffer := make([]byte, 1024)
			count, err := errorPipe.Read(buffer)
			if err != nil {
				return err
			}

			return errors.New("Could not mount VFS: " + string(buffer[0:count]))
		}
	}

	return nil
}
