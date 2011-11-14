package main

import (
	"errors"
)

type VfsCommand struct{}

func (VfsCommand) Name() string {
	return "vfs"
}

func (VfsCommand) Summary() string {
	return ""
}

func (VfsCommand) Help() string {
	return ""
}

func (VfsCommand) Exec(args []string) error {
	if len(args) == 0 {
		errors.New("Mountpoint not specified.")
	}

	database := args[0]
	mountPath := args[1]

	vfs, error := MountVfs(database, mountPath)
	if error != nil {
		return error
	}
	defer vfs.Unmount()

	vfs.Loop()

	return nil
}
