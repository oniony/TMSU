package main

import (
          "errors"
       )

type MountCommand struct {}

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
    if (len(args) == 0) { errors.New("Mountpoint not specified.") }

    mountPath := args[0]

    vfs, error := MountVfs(mountPath)
    if (error != nil) { return error }
    defer vfs.Unmount()

    vfs.Loop()

    return nil
}
