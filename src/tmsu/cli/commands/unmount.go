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
	"errors"
	"os"
	"os/exec"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/vfs"
)

type UnmountCommand struct {
	verbose bool
}

func (UnmountCommand) Name() cli.CommandName {
	return "unmount"
}

func (UnmountCommand) Synopsis() string {
	return "Unmount the virtual file-system"
}

func (UnmountCommand) Description() string {
	return `tmsu unmount MOUNTPOINT
tmsu unmount --all

Unmounts the virtual file-system at MOUNTPOINT.`
}

func (UnmountCommand) Options() cli.Options {
	return cli.Options{{"-a", "--all", "unmounts all mounted TMSU file-systems"}}
}

func (command UnmountCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	if options.HasOption("--all") {
		return command.unmountAll()
	}

	if len(args) < 1 {
		return errors.New("Path to unmount not speciified.")
	}

	return command.unmount(args[0])
}

func (command UnmountCommand) unmount(path string) error {
	if command.verbose {
		log.Info("searching path for fusermount.")
	}

	fusermountPath, err := exec.LookPath("fusermount")
	if err != nil {
		return err
	}

	if command.verbose {
		log.Infof("running: %v -u %v.", fusermountPath, path)
	}

	process, err := os.StartProcess(fusermountPath, []string{fusermountPath, "-u", path}, &os.ProcAttr{})
	if err != nil {
		return err
	}

	if command.verbose {
		log.Info("waiting for process to exit.")
	}

	processState, err := process.Wait()
	if err != nil {
		return err
	}
	if !processState.Success() {
		return errors.New("Could not unmount.")
	}

	return nil
}

func (command UnmountCommand) unmountAll() error {
	if command.verbose {
		log.Info("retrieving mount table.")
	}

	mt, err := vfs.GetMountTable()
	if err != nil {
		return errors.New("Could not get mount table: " + err.Error())
	}

	if command.verbose && len(mt) == 0 {
		log.Info("mount table is empty.")
	}

	for _, mount := range mt {
		err = command.unmount(mount.MountPath)
		if err != nil {
			return err
		}
	}

	return nil
}
