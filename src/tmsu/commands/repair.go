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
)

type RepairCommand struct{}

func (RepairCommand) Name() string {
	return "repair"
}

func (RepairCommand) Synopsis() string {
	return "Repair breakages caused by file moves and amendments"
}

func (RepairCommand) Description() string {
	return `tmsu repair

Attempts to repair the database with respect to changes to tagged file contents
and file moves.

This process consists of two steps:

  1. Update the stored fingerprints for modified files.
  2. Find the new path of moved files.`
}

func (command RepairCommand) Exec(args []string) error {
    return nil
}
