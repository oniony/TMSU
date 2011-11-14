package main

import (
	"errors"
	"os"
	"exec"
)

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

	//TODO support for explicit database

	mountPath := args[0]
	command := exec.Command(os.Args[0], "vfs", databasePath(), mountPath)

	error := command.Start()
	if error != nil {
		return error
	}

	return nil
}
