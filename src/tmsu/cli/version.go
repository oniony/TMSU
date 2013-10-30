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
	"tmsu/common"
)

var VersionCommand = Command{
	Name:     "version",
	Synopsis: "Display version and copyright information",
	Description: `tmsu version

Displays version and copyright information.`,
	Options: Options{},
	Exec:    versionExec,
}

func versionExec(options Options, args []string) error {
	fmt.Print("TMSU", common.Version)
	fmt.Print()
	fmt.Print(`Copyright © 2011-2013 Paul Ruane.

This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under certain conditions.
See the accompanying COPYING file for further details.`)

	return nil
}
