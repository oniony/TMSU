/*
Copyright 2011-2013 Paul Ruane.

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
	"fmt"
	"tmsu/cli"
	"tmsu/vfs"
)

type VfsCommand struct{}

func (VfsCommand) Name() cli.CommandName {
	return "vfs"
}

func (VfsCommand) Synopsis() string {
	return ""
}

func (VfsCommand) Description() string {
	return ""
}

func (VfsCommand) Options() cli.Options {
	return cli.Options{{"--allow-other", "-o", "turn on FUSE 'allow_other' option"}}
}

func (VfsCommand) Exec(options cli.Options, args []string) error {
	if len(args) == 0 {
		fmt.Errorf("Mountpoint not specified.")
	}

	allowOther := options.HasOption("--allow-other")
	databasePath := args[0]
	mountPath := args[1]

	vfs, err := vfs.MountVfs(databasePath, mountPath, allowOther)
	if err != nil {
		return fmt.Errorf("could not mount virtual filesystem for database '%v' at '%v': %v", databasePath, mountPath, err)
	}
	defer vfs.Unmount()

	vfs.Loop()

	return nil
}
