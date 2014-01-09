/*
Copyright 2011-2014 Paul Ruane.

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

package cli

import (
	"fmt"
	"strings"
	"tmsu/vfs"
)

var VfsCommand = Command{
	Name:     "vfs",
	Synopsis: "",
	Description: `This command is the foreground process which hosts the virtual filesystem.
It is run automatically when a virtual filesystem is mounted using the 'mount' command
and terminated when the virtual filesystem is unmounted.

It is not normally necessary to issue this command manually unless debugging the virtual
filesystem. For debug output use the --verbose option.`,
	Options: Options{{"--options", "-o", "mount options", true, ""}},
	Exec:    vfsExec,
}

func vfsExec(options Options, args []string) error {
	if len(args) == 0 {
		fmt.Errorf("Mountpoint not specified.")
	}

	mountOptions := []string{}
	if options.HasOption("--options") {
		mountOptions = strings.Split(options.Get("--options").Argument, ",")
	}

	databasePath := args[0]
	mountPath := args[1]

	vfs, err := vfs.MountVfs(databasePath, mountPath, mountOptions)
	if err != nil {
		return fmt.Errorf("could not mount virtual filesystem for database '%v' at '%v': %v", databasePath, mountPath, err)
	}
	defer vfs.Unmount()

	vfs.Serve()

	return nil
}
