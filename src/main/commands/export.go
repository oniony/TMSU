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
	"errors"
	"fmt"
	"strings"
)

type ExportCommand struct{}

func (this ExportCommand) Name() string {
	return "export"
}

func (this ExportCommand) Summary() string {
	return "exports the tag database"
}

func (this ExportCommand) Help() string {
	return `tmsu export
        
dumps the tag database to standard output as comma-separated values (CSV)`
}

func (this ExportCommand) Exec(args []string) error {
	if len(args) != 0 {
		return errors.New("Unpected argument to command '" + this.Name() + "'.")
	}

	db, error := OpenDatabase(databasePath())
	if error != nil {
		return error
	}
	defer db.Close()

	files, error := db.Files()
	if error != nil {
		return error
	}

	for _, file := range *files {
		fmt.Printf("%v,%v,", file.Path(), file.Fingerprint)

		tags, error := db.TagsByFileId(file.Id)
		if error != nil {
			return error
		}

		tagNames := make([]string, 0, len(tags))

		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}

		fmt.Println(strings.Join(tagNames, ","))
	}

	return nil
}
