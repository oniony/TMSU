/*
Copyright 2011-2015 Paul Ruane.

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
	"strings"
	"tmsu/entities"
	"tmsu/storage"
)

var ConfigCommand = Command{
	Name:     "config",
	Synopsis: "Views or amends database settings",
	Usages: []string{"tmsu config",
		"tmsu config NAME...",
	    "tmsu config NAME=VALUE..."},
	Description: `Lists or views the database settings for the current database.

Without arguments the complete set of settings are shown, otherwise lists the settings for the specified setting NAMEs.

If a VALUE is specified then the setting is updated.`,
	Options: Options{},
	Exec: configExec,
}

func configExec(store *storage.Storage, options Options, args []string) error {
    if len(args) == 0 {
        listSettings(store)
    }

    for _, arg := range args {
        parts := strings.Split(arg, "=")
        switch len(parts) {
        case 1:
            name := parts[0]
            if err := listSetting(store, name); err != nil {
                fmt.Errorf("could not show value for setting '%v': %v", name, err)
            }
        case 2:
            name := parts[0]
            value := parts[1]
            if err := amendSetting(store, name, value); err != nil {
                fmt.Errorf("could not amend setting '%v' to '%v': %v", name, value, err)
            }
        default:
            return fmt.Errorf("invalid argument, '%v'", arg)
        }
    }

    return nil
}

// unexported

func listSettings(store *storage.Storage) error {
    settings, err := store.Settings()
    if err != nil {
        return fmt.Errorf("could not retrieve settings: %v", err)
    }

    for _, setting := range settings {
        printSetting(setting)
    }

    return nil
}

func listSetting(store *storage.Storage, name string) error {
    setting, err := store.Setting(name)
    if err != nil {
        return fmt.Errorf("could not retrieve setting '%v'", err)
    }
    if setting == nil {
        return fmt.Errorf("no such setting '%v'", name)
    }

    printSetting(setting)

    return nil
}

func amendSetting(store *storage.Storage, name, value string) error {
    setting, err := store.Setting(name)
    if err != nil {
        return fmt.Errorf("could not retrieve setting '%v'", err)
    }
    if setting == nil {
        return fmt.Errorf("no such setting '%v'", name)
    }

    if _, err = store.UpdateSetting(name, value); err != nil {
        return fmt.Errorf("could not update setting '%v': %v", name, err)
    }

    return nil
}

func printSetting(setting *entities.Setting) {
    fmt.Printf("%v=%v\n", setting.Name, setting.Value)
}
