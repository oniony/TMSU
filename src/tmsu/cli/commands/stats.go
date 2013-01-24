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

package commands

import (
	"fmt"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type StatsCommand struct{}

func (StatsCommand) Name() cli.CommandName {
	return "stats"
}

func (StatsCommand) Synopsis() string {
	return "Show database statistics"
}

func (StatsCommand) Description() string {
	return `tmsu stats

Shows the database statistics.`
}

func (StatsCommand) Options() cli.Options {
	return cli.Options{}
}

func (StatsCommand) Exec(options cli.Options, args []string) error {
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

	explicitFileTagCount, err := store.ExplicitFileTagCount()
	if err != nil {
		return fmt.Errorf("could not retrieve explicit taggings count: %v", err)
	}

	implicitFileTagCount, err := store.ImplicitFileTagCount()
	if err != nil {
		return fmt.Errorf("could not retrieve implicit taggings count: %v", err)
	}

	fileTagCount := explicitFileTagCount + implicitFileTagCount

	log.Printf("Database Contents")

	log.Printf(" Tags:        %v", tagCount)
	log.Printf(" Files:       %v", fileCount)
	log.Print()
	log.Printf(" Taggings     total")
	log.Printf("   all:       %v", fileTagCount)
	log.Printf("   explicit:  %v", explicitFileTagCount)
	log.Printf("   implicit:  %v", implicitFileTagCount)

	return nil
}
