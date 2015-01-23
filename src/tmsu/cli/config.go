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
	Exec:    configExec,
}

func configExec(store *storage.Storage, options Options, args []string) error {
	if len(args) == 0 {
		if err := listAllSettings(store); err != nil {
			return fmt.Errorf("could not list settings")
		}
	}

	if len(args) == 1 && strings.Index(args[0], "=") == -1 {
		printSettingValue(store, args[0])
		return nil
	}

	for _, arg := range args {
		parts := strings.Split(arg, "=")
		switch len(parts) {
		case 1:
			name := parts[0]
			if err := printSetting(store, name); err != nil {
				return fmt.Errorf("could not show value for setting '%v': %v", name, err)
			}
		case 2:
			name := parts[0]
			value := parts[1]

			if err := amendSetting(store, name, value); err != nil {
				return fmt.Errorf("could not amend setting '%v' to '%v': %v", name, value, err)
			}
		default:
			return fmt.Errorf("invalid argument, '%v'", arg)
		}
	}

	return nil
}

// unexported

func listAllSettings(store *storage.Storage) error {
	settings, err := store.Settings()
	if err != nil {
		return fmt.Errorf("could not retrieve settings: %v", err)
	}

	width := 0
	for _, setting := range settings {
		if len(setting.Name) > width {
			width = len(setting.Name)
		}
	}

	for _, setting := range settings {
		fmt.Printf("%*v %v\n", -width, setting.Name, setting.Value)
	}

	return nil
}

func printSetting(store *storage.Storage, name string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}

	setting, err := store.Setting(name)
	if err != nil {
		return fmt.Errorf("could not retrieve setting '%v'", err)
	}
	if setting == nil {
		return fmt.Errorf("no such setting '%v'", name)
	}

	fmt.Printf("%v %v\n", setting.Name, setting.Value)

	return nil
}

func printSettingValue(store *storage.Storage, name string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}

	setting, err := store.Setting(name)
	if err != nil {
		return fmt.Errorf("could not retrieve setting '%v'", err)
	}
	if setting == nil {
		return fmt.Errorf("no such setting '%v'", name)
	}

	fmt.Println(setting.Value)

	return nil
}

func amendSetting(store *storage.Storage, name, value string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}
	if value == "" {
		return fmt.Errorf("setting '%v' value must be specified", name)
	}

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
