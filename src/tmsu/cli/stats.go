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
	"tmsu/log"
	"tmsu/storage"
)

var StatsCommand = &Command{
	Name:     "stats",
	Synopsis: "Show database statistics",
	Description: `tmsu stats

Shows the database statistics.`,
	Options: Options{},
	Exec:    statsExec,
}

func statsExec(options Options, args []string) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	tagCount, err := store.TagCount()
	if err != nil {
		return fmt.Errorf("could not retrieve tag count: %v", err)
	}

	fileCount, err := store.FileCount()
	if err != nil {
		return fmt.Errorf("could not retrieve file count: %v", err)
	}

	fileTagCount, err := store.FileTagCount()
	if err != nil {
		return fmt.Errorf("could not retrieve taggings count: %v", err)
	}

	log.Printf("     Tags: %v", tagCount)
	log.Printf("    Files: %v", fileCount)
	log.Printf(" Taggings: %v", fileTagCount)

	return nil
}
