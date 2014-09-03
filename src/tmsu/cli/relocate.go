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
	"tmsu/storage"
)

var RelocateCommand = Command{
	Name:     "relocate",
	Synopsis: "Update the paths of moved files or directories",
	Description: `tmsu relocate PATH NEW

Updates the paths in the database for all files under PATH.

Examples:

    $ tmsu relocate /some/file /new/location
    $ tmsu relocate / /mnt/ext`,
	Options: Options{},
	Exec:    relocateExec,
}

func relocateExec(options Options, args []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if err := store.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer store.Commit()

	//TODO implement
	//TODO if verbose, log old and new path for each moved file

	wereErrors := false

	if wereErrors {
		return errBlank
	}

	return nil
}
