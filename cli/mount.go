// Copyright 2011-2018 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// +build !windows

package cli

import (
	"fmt"
	"github.com/oniony/TMSU/common/log"
	"github.com/oniony/TMSU/vfs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

var MountCommand = Command{
	Name:     "mount",
	Synopsis: "Mount the virtual filesystem",
	Usages: []string{"tmsu mount",
		"tmsu mount [OPTION]... MOUNTPOINT"},
	Description: `Without arguments, lists the currently mounted file-systems, otherwise mounts a virtual file-system at the path MOUNTPOINT.

Use the --database global option or TMSU_DB environment variable (in order of precedence) to specify the path of the database to be mounted.

Where neither --database is specified nor TMSU_DB defined then the default database is mounted.

To allow other users access to the mounted filesystem, pass the 'allow_other' FUSE option, e.g. 'tmsu mount --options=allow_other mp'. (FUSE only allows the root user to use this option unless 'user_allow_other' is present in '/etc/fuse.conf'.)

For further documentation on the usage of the --database option, refer to 'tmsu help', without specifying a subcommand`,
	Examples: []string{"$ tmsu mount mp",
		"$ tmsu mount --database=/tmp/db mp",
		"$ tmsu mount --options=allow_other mp"},
	Options: Options{Option{"--options", "-o", "mount options (passed to fusermount)", true, ""}},
	Exec:    mountExec,
}

// unexported

func mountExec(options Options, args []string, databasePath string) (error, warnings) {
	var mountOptions string
	if options.HasOption("--options") {
		mountOptions = options.Get("--options").Argument
	}

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	switch len(args) {
	case 0:
		if err := listMounts(); err != nil {
			return err, nil
		}
	case 1:
		mountPath := args[0]

		if err := mountExplicit(store.DbPath, mountPath, mountOptions); err != nil {
			return err, nil
		}
	case 2:
		databasePath := args[0]
		mountPath := args[1]

		if err := mountExplicit(databasePath, mountPath, mountOptions); err != nil {
			return err, nil
		}
	default:
		return fmt.Errorf("too many arguments"), nil
	}

	return nil, nil
}

func listMounts() error {
	log.Info(2, "retrieving mount table.")

	mt, err := vfs.GetMountTable()
	if err != nil {
		return fmt.Errorf("could not get mount table: %v", err)
	}

	if len(mt) == 0 {
		log.Info(2, "mount table is empty.")
	}

	dbPathWidth := 0
	for _, mount := range mt {
		if len(mount.DatabasePath) > dbPathWidth {
			dbPathWidth = len(mount.DatabasePath)
		}
	}

	for _, mount := range mt {
		fmt.Printf("%-*v\tat\t%v\n", dbPathWidth, mount.DatabasePath, mount.MountPath)
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
		return fmt.Errorf("%v: mount point does not exist", mountPath)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%v: mount point is not a directory", mountPath)
	}

	stat, err = os.Stat(databasePath)
	if err != nil {
		return fmt.Errorf("%v: could not stat: %v", databasePath, err)
	}
	if stat == nil {
		return fmt.Errorf("%v: database does not exist", databasePath)
	}

	log.Infof(2, "spawning daemon to mount VFS for database '%v' at '%v'", databasePath, mountPath)

	args := []string{"vfs", "--database=" + databasePath, mountPath, "--options=" + mountOptions}
	daemon := exec.Command(os.Args[0], args...)

	tempFile, err := ioutil.TempFile("", "tmsu-vfs-")
	if err != nil {
		return fmt.Errorf("could not get a temporary file: %v", err)
	}
	daemon.Stderr = tempFile

	err = daemon.Start()
	if err != nil {
		return fmt.Errorf("could not start daemon: %v", err)
	}

	log.Info(2, "sleeping.")

	const halfSecond = 500000000
	time.Sleep(halfSecond)

	log.Info(2, "checking whether daemon started successfully.")

	var waitStatus syscall.WaitStatus
	var rusage syscall.Rusage
	_, err = syscall.Wait4(daemon.Process.Pid, &waitStatus, syscall.WNOHANG, &rusage)
	if err != nil {
		return fmt.Errorf("could not check daemon status: %v", err)
	}

	if waitStatus.Exited() {
		if waitStatus.ExitStatus() != 0 {
			return fmt.Errorf("virtual filesystem mount failed: see standard error output: %v", tempFile.Name())
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
