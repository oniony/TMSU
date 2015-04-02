// Copyright 2011-2015 Paul Ruane.

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
	"math"
	"tmsu/common/terminal/ansi"
	"tmsu/storage"
)

var InfoCommand = Command{
	Name:        "info",
	Synopsis:    "Show database information",
	Usages:      []string{"tmsu info"},
	Description: "Shows the database information.",
	Options: Options{
		Option{"--stats", "-s", "show statistics", false, ""},
		Option{"--usage", "-u", "show tag usage breakdown", false, ""}},
	Exec:    infoExec,
	Aliases: []string{"stats"},
}

func infoExec(store *storage.Storage, options Options, args []string) error {
	stats := options.HasOption("--stats")
	usage := options.HasOption("--usage")
	colour, err := useColour(options)
	if err != nil {
		return err
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	printInfo("Database", store.DbPath, colour)
	printInfo("Root path", store.RootPath, colour)

	if stats {
		showStatistics(store, tx, colour)
	}
	if usage {
		showUsage(store, tx, colour)
	}

	return nil
}

func printInfo(name string, value string, colour bool) {
	if colour {
		value = ansi.Green(value)
	}

	fmt.Printf("%v: %v\n", name, value)
}

func showStatistics(store *storage.Storage, tx *storage.Tx, colour bool) error {
	tagCount, err := store.TagCount(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve tag count: %v", err)
	}

	valueCount, err := store.ValueCount(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve value count: %v", err)
	}

	fileCount, err := store.FileCount(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve file count: %v", err)
	}

	fileTagCount, err := store.FileTagCount(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve taggings count: %v", err)
	}

	var averageTagsPerFile float32
	if fileCount > 0 {
		averageTagsPerFile = float32(fileTagCount) / float32(fileCount)
	}

	var averageFilesPerTag float32
	if tagCount > 0 {
		averageFilesPerTag = float32(fileTagCount) / float32(tagCount)
	}

	fmt.Println()
	fmt.Printf("Tags: %v\n", tagCount)
	fmt.Printf("Values: %v\n", valueCount)
	fmt.Printf("Files: %v\n", fileCount)
	fmt.Printf("Taggings: %v\n", fileTagCount)
	fmt.Println()

	fmt.Printf("Mean tags per file: %1.2f\n", averageTagsPerFile)
	fmt.Printf("Mean files per tag: %1.2f\n", averageFilesPerTag)

	return nil
}

func showUsage(store *storage.Storage, tx *storage.Tx, colour bool) error {
	tagUsages, err := store.TagUsage(tx)
	if err != nil {
		return fmt.Errorf("could not retrieve tag usage: %v", err)
	}

	maxLength := 0
	maxCountWidth := 0

	for _, tagUsage := range tagUsages {
		countWidth := int(math.Log(float64(tagUsage.FileCount)))
		if countWidth > maxCountWidth {
			maxCountWidth = countWidth
		}
		if len(tagUsage.Name) > maxLength {
			maxLength = len(tagUsage.Name)
		}
	}

	fmt.Println()
	for _, tagUsage := range tagUsages {
		fmt.Printf("  %*s %*v\n", -maxLength, tagUsage.Name, maxCountWidth, tagUsage.FileCount)
	}

	return nil
}
