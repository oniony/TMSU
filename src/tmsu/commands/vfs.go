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

package commands

import (
	"errors"
	"tmsu/vfs"
)

type VfsCommand struct{}

func (VfsCommand) Name() string {
	return "vfs"
}

func (VfsCommand) Synopsis() string {
	return ""
}

func (VfsCommand) Description() string {
	return ""
}

func (VfsCommand) Exec(args []string) error {
	if len(args) == 0 {
		errors.New("Mountpoint not specified.")
	}

	database := args[0]
	mountPath := args[1]

	vfs, err := vfs.MountVfs(database, mountPath)
	if err != nil {
		return err
	}
	defer vfs.Unmount()

	vfs.Loop()

	return nil
}
