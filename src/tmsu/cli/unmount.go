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
	"tmsu/log"
	"tmsu/vfs"
)

var UnmountCommand = Command{
	Name:     "unmount",
	Synopsis: "Unmount the virtual filesystem",
	Description: `tmsu unmount MOUNTPOINT
tmsu unmount --all

Unmounts the virtual file-system at MOUNTPOINT.`,
	Options: Options{{"--all", "-a", "unmounts all mounted TMSU file-systems", false, ""}},
	Exec:    unmountExec,
}

func unmountExec(options Options, args []string) error {
	if options.HasOption("--all") {
		return unmountAll()
	}

	if len(args) < 1 {
		return fmt.Errorf("path to unmount not speciified.")
	}

	return unmount(args[0])
}

func unmount(path string) error {
	log.Supp("searching path for fusermount.")

	fusermountPath, err := exec.LookPath("fusermount")
	if err != nil {
		return fmt.Errorf("could not find 'fusermount': ensure fuse is installed: %v", err)
	}

	log.Suppf("running: %v -u %v.", fusermountPath, path)

	process, err := os.StartProcess(fusermountPath, []string{fusermountPath, "-u", path}, &os.ProcAttr{})
	if err != nil {
		return fmt.Errorf("could not start 'fusermount': %v", err)
	}

	log.Supp("waiting for process to exit.")

	processState, err := process.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for process to exit: %v", err)
	}
	if !processState.Success() {
		return fmt.Errorf("could not unmount virtual filesystem.")
	}

	return nil
}

func unmountAll() error {
	log.Supp("retrieving mount table.")

	mt, err := vfs.GetMountTable()
	if err != nil {
		return fmt.Errorf("could not get mount table: %v", err)
	}

	if len(mt) == 0 {
		log.Supp("mount table is empty.")
	}

	for _, mount := range mt {
		err = unmount(mount.MountPath)
		if err != nil {
			return err
		}
	}

	return nil
}
