package main

import (
          "errors"
          "log"
       )

type MountCommand struct {}

func (this MountCommand) Name() string {
    return "mount"
}

func (this MountCommand) Description() string {
    return "mounts the virtual file-system"
}

func (this MountCommand) Exec(args []string) error {
    if (len(args) == 0) { errors.New("Mountpoint not specified.") }

    mountPath := args[0]

    vfs, error := MountVfs(mountPath)
    if (error != nil) { return error }
    defer vfs.Unmount()

    log.Printf("Database '%v' mounted at '%v'.\n", databasePath(), mountPath)

    vfs.Loop()

    return nil
}
