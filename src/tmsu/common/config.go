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

package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

const globalConfigPath = "/etc/tmsu.conf"
const userConfigPath = "~/.config/tmsu.conf"
const defaultDatabasePath = "~/.tmsu/default.db"

type DatabaseConfig struct {
	Name         string
	DatabasePath string
}

func GetSelectedDatabaseConfig() (*DatabaseConfig, error) {
	path := os.Getenv("TMSU_DB")
	if path == "" {
		return nil, nil
	}

	return &DatabaseConfig{"", path}, nil
}

func GetDefaultDatabaseConfig() (*DatabaseConfig, error) {
	path, err := resolvePath(defaultDatabasePath)
	if err != nil {
		return nil, errors.New("Could not resolve default database path: " + err.Error())
	}

	return &DatabaseConfig{"default", path}, nil
}

func resolvePath(path string) (string, error) {
	if strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		user, err := user.Current()
		if err != nil {
			return "", errors.New("Could not identify home directory: " + err.Error())
		}

		path = strings.Join([]string{user.HomeDir, path[2:]}, string(filepath.Separator))
	}

	return path, nil
}

func readConfig(path string) ([]DatabaseConfig, error) {
	configPath, err := resolvePath(path)
	if err != nil {
		return nil, errors.New("Could not resolve configuration file path '" + path + "'.")
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // configuration file is missing
		} else if os.IsPermission(err) {
			return nil, errors.New("Permission denied.")
		} else {
			return nil, err
		}
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	databases := make([]DatabaseConfig, 0, 5)
	var database *DatabaseConfig

	for lineBytes, _, err := reader.ReadLine(); err == nil; lineBytes, _, err = reader.ReadLine() {
		line := string(lineBytes)
		trimmedLine := strings.TrimLeft(line, " \t")

		if len(trimmedLine) == 0 {
			continue
		}
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		if line[0] != ' ' && line[0] != '\t' && database != nil {
			databases = append(databases, *database)
			database = nil
		}

		var name, quotedValue string
		count, err := fmt.Sscanf(trimmedLine, "%s %s", &name, &quotedValue)
		if count < 2 {
			return nil, errors.New("Key and value must be specified.")
		}
		if err != nil {
			return nil, err
		}

		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			return nil, errors.New("Configuration error: values must be quoted.")
		}

		switch name {
		case "database":
			database = &DatabaseConfig{}
			database.Name = value
			if err != nil {
				return nil, err
			}
		case "path":
			path, err := resolvePath(value)
			if err != nil {
				return nil, err
			}

			database.DatabasePath = path
		default:
			return nil, errors.New("Unrecognised configuration element name '" + name + "'.")
		}
	}

	if database != nil {
		databases = append(databases, *database)
	}

	return databases, nil
}
