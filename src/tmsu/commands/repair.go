/*
Copyright 2011-2012 Paul Ruane.

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
//	"tmsu/fingerprint"
)

type RepairCommand struct{}

func (RepairCommand) Name() string {
	return "repair"
}

func (RepairCommand) Synopsis() string {
	return "Repair the database"
}

func (RepairCommand) Description() string {
	return `tmsu repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

Repairs tagged files and directories under PATHS by:

    1. Updating the stored fingerprints of modified files and directories.
    2. Updating the path of missing files and directories where an untagged file
       or directory with the same fingerprint can be found in PATHs.

Where no PATHS are specified all tagged files and directories fingerprints in
the database are repaired. (In this mode file moves cannot be repaired as tmsu
will not know where to look for them.)`
}

func (command RepairCommand) Exec(args []string) error {
    //pathsByFingerprint := make(map[fingerprint.Fingerprint]string)

    //for _, path := range args {
        //TODO recursively build fingerprints
   // }

    //TODO get the database entries under PATHs
    //TODO look at file in file system
    //TODO   case missing: is there an untagged file with same fingerprint?
    //         case more: warn multiple candidates
    //         case 1: update database
    //         case 0: warn file missing
    //TODO   case modified: update database
    //TODO   case unmodified: do nothing

	return nil
}
