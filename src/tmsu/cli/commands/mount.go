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
	"os"
	"os/exec"
	"syscall"
	"time"
	"tmsu/cli"
	"tmsu/common"
	"tmsu/log"
	"tmsu/vfs"
)

type MountCommand struct {
	verbose    bool
	allowOther bool
}

func (MountCommand) Name() cli.CommandName {
	return "mount"
}

func (MountCommand) Synopsis() string {
	return "Mount the virtual file-system"
}

func (MountCommand) Description() string {
	return `tmsu mount
tmsu mount [FILE] MOUNTPOINT

Without arguments, lists the currently mounted file-systems, otherwise mounts a
virtual file-system at the path MOUNTPOINT.

Where FILE is specified, the database at FILE is mounted.

If FILE is not specified but the TMSU_DB environment variable is defined then
the database at TMSU_DB is mounted.

Where neither FILE is specified nor TMSU_DB defined then the default database
is mounted.`
}

func (MountCommand) Options() cli.Options {
	return cli.Options{{"--allow-other", "-o", "allow other users access to the VFS (requires root or setting in fuse.conf)", false, ""}}
}

func (command MountCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")
	command.allowOther = options.HasOption("--allow-other")

	argCount := len(args)

	switch argCount {
	case 0:
		err := command.listMounts()
		if err != nil {
			return fmt.Errorf("could not list mounts: %v", err)
		}
	case 1:
		mountPath := args[0]

		err := command.mountSelected(mountPath)
		if err != nil {
			return fmt.Errorf("could not mount database at '%v': %v", mountPath, err)
		}
	case 2:
		databasePath := args[0]
		mountPath := args[1]

		err := command.mountExplicit(databasePath, mountPath)
		if err != nil {
			return fmt.Errorf("could not mount database '%v' at '%v': %v", databasePath, mountPath, err)
		}
	default:
		return fmt.Errorf("Too many arguments.")
	}

	return nil
}

func (command MountCommand) listMounts() error {
	if command.verbose {
		log.Info("retrieving mount table.")
	}

	mt, err := vfs.GetMountTable()
	if err != nil {
		return fmt.Errorf("could not get mount table: %v", err)
	}

	if command.verbose && len(mt) == 0 {
		log.Info("mount table is empty.")
	}

	for _, mount := range mt {
		log.Printf("'%v' at '%v'", mount.DatabasePath, mount.MountPath)
	}

	return nil
}

func (command MountCommand) mountSelected(mountPath string) error {
	databasePath, err := common.GetDatabasePath()
	if err != nil {
		return fmt.Errorf("could not get selected database configuration: %v", err)
	}

	if err = command.mountExplicit(databasePath, mountPath); err != nil {
		return err
	}

	return nil
}

func (command MountCommand) mountExplicit(databasePath string, mountPath string) error {
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

	if command.verbose {
		log.Infof("spawning daemon to mount VFS for database '%v' at '%v'.", databasePath, mountPath)
	}

	args := []string{"vfs", databasePath, mountPath}
	if command.allowOther {
		args = append(args, "--allow-other")
	}

	daemon := exec.Command(os.Args[0], args...)

	errorPipe, err := daemon.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not open standard error pipe: %v", err)
	}

	err = daemon.Start()
	if err != nil {
		return fmt.Errorf("could not start daemon: %v", err)
	}

	if command.verbose {
		log.Info("sleeping.")
	}

	const HALF_SECOND = 500000000
	time.Sleep(HALF_SECOND)

	if command.verbose {
		log.Info("checking whether daemon started successfully.")
	}

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
