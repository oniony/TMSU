// Copyright 2011 Paul Ruane. All rights reserved.

package main

import (
	"errors"
	"exec"
	"os"
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
	if len(args) < 1 {
		errors.New("No mountpoint specified.")
	}
	if len(args) > 1 {
		errors.New("Extraneous arguments.")
	}

    //TODO check the mount-point exists
    //TODO check the mount-point permissions
    //TODO check the database exists
    //TODO check the database permissions

	mountPath := args[0]
	command := exec.Command(os.Args[0], "vfs", databasePath(), mountPath)

	errorPipe, error := command.StderrPipe()
	if error != nil {
	    return error
    }

	error = command.Start()
	if error != nil {
		return error
	}

    time.Sleep(HALF_SECOND)

    waitMessage, error := command.Process.Wait(os.WNOHANG)
    if error != nil {
        return error
    }

    if waitMessage.WaitStatus.Exited() {
        if waitMessage.WaitStatus.ExitStatus() != 0 {
            buffer := make([]byte, 1024)
            count, error := errorPipe.Read(buffer)
            if error != nil {
                return error
            }

            return errors.New("Could not mount VFS: " + string(buffer[0:count]))
        }
    }

	return nil
}
