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
	"fmt"
	"tmsu/common"
)

type VersionCommand struct{}

func (VersionCommand) Name() string {
	return "version"
}

func (VersionCommand) Synopsis() string {
	return "Display version and copyright information"
}

func (VersionCommand) Description() string {
	return `tmsu version

Displays version and copyright information.`
}

func (VersionCommand) Exec(args []string) error {
	fmt.Println("TMSU", common.Version)
	fmt.Println()
	fmt.Println(`Copyright © 2011 Paul Ruane.

This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under certain conditions.
See the accompanying LICENSE file for further details.`)

	return nil
}
