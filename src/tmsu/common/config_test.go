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

package common

import (
	"os"
	"testing"
)

func TestDefaultDatabase(test *testing.T) {
    os.Clearenv()
    os.Setenv("HOME", "/foo/bar/home")

    config, err := GetDefaultDatabaseConfig()
    if err != nil {
        test.Fatal(err.Error())
    }
    if config == nil {
        test.Fatal("Could not retrieve default database configuration.")
    }
    if config.Name != "default" {
        test.Fatal("Default database has incorrect name '" + config.Name + "'.")
    }
    if config.DatabasePath != "/foo/bar/home/.tmsu/default.db" {
        test.Fatal("Default database config has incorrect path '" + config.DatabasePath + "'.")
    }
}

func TestSelectedDatabaseWhenEnvVarUndefined(test *testing.T) {
    os.Clearenv()

    config, err := GetSelectedDatabaseConfig()
    if err != nil {
        test.Fatal(err.Error())
    }
    if config != nil {
        test.Fatal("Expected selected database configuration to be nil.")
    }
}

func TestSelectedDatabaseWhenEnvVarDefined(test *testing.T) {
    os.Clearenv()
    os.Setenv("TMSU_DB", "/some/where/over/the/rainbow")

    config, err := GetSelectedDatabaseConfig()
    if err != nil {
        test.Fatal(err.Error())
    }
    if config == nil {
        test.Fatal("Could not retrieve database configuration.")
    }
    if config.DatabasePath != "/some/where/over/the/rainbow" {
        test.Fatal("Selected database different to that expected.")
    }
}
