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

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"tmsu/log"
	"tmsu/storage/database"
	"tmsu/vfs"
)

var MountCommand = Command{
	Name:     "mount",
	Synopsis: "Mount the virtual filesystem",
	Description: `tmsu mount
tmsu mount [OPTION]... [FILE] MOUNTPOINT

Without arguments, lists the currently mounted file-systems, otherwise mounts a
virtual file-system at the path MOUNTPOINT.

Where FILE is specified, the database at FILE is mounted.

If FILE is not specified but the TMSU_DB environment variable is defined then
the database at TMSU_DB is mounted.

Where neither FILE is specified nor TMSU_DB defined then the default database
is mounted.

To allow other users access to the mounted filesystem, pass the 'allow_other'
FUSE option, e.g. 'tmsu mount --option=allow_other mp'. (FUSE only allows the
root user to use this option unless 'user_allow_other' is present in
'/etc/fuse.conf'.)`,
	Options: Options{Option{"--options", "-o", "mount options (passed to fusermount)", true, ""}},
	Exec:    mountExec,
}

func mountExec(options Options, args []string) error {
	var mountOptions string
	if options.HasOption("--options") {
		mountOptions = options.Get("--options").Argument
	}

	argCount := len(args)

	switch argCount {
	case 0:
		err := listMounts()
		if err != nil {
			return err
		}
	case 1:
		mountPath := args[0]

		err := mountDefault(mountPath, mountOptions)
		if err != nil {
			return err
		}
	case 2:
		databasePath := args[0]
		mountPath := args[1]

		err := mountExplicit(databasePath, mountPath, mountOptions)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Too many arguments.")
	}

	return nil
}

func listMounts() error {
	log.Supp("retrieving mount table.")

	mt, err := vfs.GetMountTable()
	if err != nil {
		return fmt.Errorf("could not get mount table: %v", err)
	}

	if len(mt) == 0 {
		log.Supp("mount table is empty.")
	}

	for _, mount := range mt {
		log.Printf("'%v' at '%v'", mount.DatabasePath, mount.MountPath)
	}

	return nil
}

func mountDefault(mountPath string, mountOptions string) error {
	if err := mountExplicit(database.Path, mountPath, mountOptions); err != nil {
		return err
	}

	return nil
}

func mountExplicit(databasePath string, mountPath string, mountOptions string) error {
	if alreadyMounted(mountPath) {
		return fmt.Errorf("%v: mount path already in use", mountPath)
	}

	stat, err := os.Stat(mountPath)
	if err != nil {
		return fmt.Errorf("%v: could not stat: %v", mountPath, err)
	}
	if stat == nil {
		return fmt.Errorf("%v: mount point does not exist.", mountPath)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%v: mount point is not a directory.", mountPath)
	}

	stat, err = os.Stat(databasePath)
	if err != nil {
		return fmt.Errorf("%v: could not stat: %v", databasePath, err)
	}
	if stat == nil {
		return fmt.Errorf("%v: database does not exist.")
	}

	log.Suppf("spawning daemon to mount VFS for database '%v' at '%v'.", databasePath, mountPath)

	args := []string{"vfs", databasePath, mountPath, "--options=" + mountOptions}
	daemon := exec.Command(os.Args[0], args...)

	errorPipe, err := daemon.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not open standard error pipe: %v", err)
	}

	err = daemon.Start()
	if err != nil {
		return fmt.Errorf("could not start daemon: %v", err)
	}

	log.Supp("sleeping.")

	const HALF_SECOND = 500000000
	time.Sleep(HALF_SECOND)

	log.Supp("checking whether daemon started successfully.")

	var waitStatus syscall.WaitStatus
	var rusage syscall.Rusage
	_, err = syscall.Wait4(daemon.Process.Pid, &waitStatus, syscall.WNOHANG, &rusage)
	if err != nil {
		return fmt.Errorf("could not check daemon status: %v", err)
	}

	if waitStatus.Exited() {
		if waitStatus.ExitStatus() != 0 {
			buffer := make([]byte, 1024)
			count, err := errorPipe.Read(buffer)
			if err != nil {
				return fmt.Errorf("could not read from error pipe: %v", err)
			}

			return fmt.Errorf("virtual filesystem mount failed: %v", string(buffer[0:count]))
		}
	}

	return nil
}

func alreadyMounted(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	mt, err := vfs.GetMountTable()
	if err != nil {
		return false
	}

	for _, mount := range mt {
		if mount.MountPath == absPath {
			return true
		}
	}

	return false
}
