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

package main

import (
	"os"
	"path/filepath"
)

const globalConfigPath = "/etc/tmsu.conf"
const userConfigPath = "~/.config/tmsu.conf"

struct DatabaseConfig {
    DatabasePath string
    MountPath string
}

func configuredDatabases() ([]DatabaseConfig, error) {
    configs := make([]DatabaseConfig, 0, 10)

    //TODO read global configuration

    //TODO read user configuration
        //TODO if not exist, create
}

func createConfigFile() error {
    defaultConfig := `# TMSU configuration file

# The default database.
database "default"
	path=~/.tmsu/default.db
	mountpoint=./tags

# An example database.
database "example"
	path=~/path/to/db
	mountpoint=~/path/to/mountpoint
    return nil`

    //TODO write to user configuration path
}
