// Copyright 2011-2017 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"github.com/oniony/TMSU/storage"
	"strings"
)

var ConfigCommand = Command{
	Name:     "config",
	Synopsis: "Views or amends database settings",
	Usages: []string{"tmsu config",
		"tmsu config NAME[=VALUE]..."},
	Description: `Lists or views the database settings for the current database.

Without arguments the complete set of settings are shown, otherwise lists the settings for the specified setting NAMEs.

If a VALUE is specified then the setting is updated.`,
	Options: Options{},
	Exec:    configExec,
}

// unexported

func configExec(options Options, args []string, databasePath string) (error, warnings) {
	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	if len(args) == 0 {
		if err := listAllSettings(store, tx); err != nil {
			return fmt.Errorf("could not list settings"), nil
		}
	}

	if len(args) == 1 && strings.Index(args[0], "=") == -1 {
		printSettingValue(store, tx, args[0])
		return nil, nil
	}

	for _, arg := range args {
		parts := strings.Split(arg, "=")
		switch len(parts) {
		case 1:
			name := parts[0]
			if err := printSetting(store, tx, name); err != nil {
				return fmt.Errorf("could not show value for setting '%v': %v", name, err), nil
			}
		case 2:
			name := parts[0]
			value := parts[1]

			if err := amendSetting(store, tx, name, value); err != nil {
				return fmt.Errorf("could not amend setting '%v' to '%v': %v", name, value, err), nil
			}
		default:
			return fmt.Errorf("invalid argument, '%v'", arg), nil
		}
	}

	return nil, nil
}

func listAllSettings(store *storage.Storage, tx *storage.Tx) error {
	settings, err := store.Settings(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve settings: %v", err)
	}

	for _, setting := range settings {
		printSettingAndValue(setting.Name, setting.Value)
	}

	return nil
}

func printSetting(store *storage.Storage, tx *storage.Tx, name string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}

	setting, err := store.Setting(tx, name)
	if err != nil {
		return fmt.Errorf("could not retrieve setting '%v'", err)
	}
	if setting == nil {
		return fmt.Errorf("no such setting '%v'", name)
	}

	printSettingAndValue(setting.Name, setting.Value)

	return nil
}

func printSettingValue(store *storage.Storage, tx *storage.Tx, name string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}

	setting, err := store.Setting(tx, name)
	if err != nil {
		return fmt.Errorf("could not retrieve setting '%v'", err)
	}
	if setting == nil {
		return fmt.Errorf("no such setting '%v'", name)
	}

	fmt.Println(setting.Value)

	return nil
}

func printSettingAndValue(name, value string) {
	fmt.Printf("%v=%v\n", name, value)
}

func amendSetting(store *storage.Storage, tx *storage.Tx, name, value string) error {
	if name == "" {
		return fmt.Errorf("setting name must be specified")
	}
	if value == "" {
		return fmt.Errorf("setting '%v' value must be specified", name)
	}

	setting, err := store.Setting(tx, name)
	if err != nil {
		return fmt.Errorf("could not retrieve setting '%v'", err)
	}
	if setting == nil {
		return fmt.Errorf("no such setting '%v'", name)
	}

	if _, err = store.UpdateSetting(tx, name, value); err != nil {
		return fmt.Errorf("could not update setting '%v': %v", name, err)
	}

	return nil
}
