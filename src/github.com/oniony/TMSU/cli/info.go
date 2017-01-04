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
	"github.com/oniony/TMSU/common/terminal/ansi"
	"github.com/oniony/TMSU/storage"
	"math"
	"os"
	"strconv"
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

// unexported

func infoExec(options Options, args []string, databasePath string) (error, warnings) {
	stats := options.HasOption("--stats")
	usage := options.HasOption("--usage")
	colour, err := useColour(options)
	if err != nil {
		return err, nil
	}

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

	showBasic(store, tx, colour)

	if stats {
		showStatistics(store, tx, colour)
	}
	if usage {
		showUsage(store, tx, colour)
	}

	return nil, nil
}

func showBasic(store *storage.Storage, tx *storage.Tx, colour bool) error {
	printInfo("Database", store.DbPath, colour)
	printInfo("Root path", store.RootPath, colour)

	stat, err := os.Stat(store.DbPath)
	if err != nil {
		return err
	}

	printInfo("Size", stat.Size(), colour)

	return nil
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
	printInfo("Tags", tagCount, colour)
	printInfo("Values", valueCount, colour)
	printInfo("Files", fileCount, colour)
	printInfo("Taggings", fileTagCount, colour)
	printInfof("Mean tags per file", "%1.2f", averageTagsPerFile, colour)
	printInfof("Mean files per tag", "%1.2f", averageFilesPerTag, colour)

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
		fileCount := strconv.FormatUint(uint64(tagUsage.FileCount), 10)
		if colour {
			fileCount = ansi.Yellow(fileCount)
		}

		fmt.Printf("  %*s %*v\n", -maxLength, tagUsage.Name, maxCountWidth, fileCount)
	}

	return nil
}

func printInfo(name string, value interface{}, colour bool) {
	printInfof(name, "%v", value, colour)
}

func printInfof(name, format string, value interface{}, colour bool) {
	if colour {
		format = ansi.Green(format)
	}

	fmt.Printf("%v: "+format+"\n", name, value)
}
