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
	"math"
	"tmsu/log"
	"tmsu/storage"
)

var StatsCommand = Command{
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

	var averageTagsPerFile float32
	if fileCount > 0 {
		averageTagsPerFile = float32(fileTagCount) / float32(fileCount)
	}

	var averageFilesPerTag float32
	if tagCount > 0 {
		averageFilesPerTag = float32(fileTagCount) / float32(tagCount)
	}

	log.Print("BASICS")
	log.Print()
	log.Printf("  Tags:     %v", tagCount)
	log.Printf("  Files:    %v", fileCount)
	log.Printf("  Taggings: %v", fileTagCount)
	log.Print()

	log.Print("AVERAGES")
	log.Print()
	log.Printf("  Tags per file: %1.2f", averageTagsPerFile)
	log.Printf("  Files per tag: %1.2f", averageFilesPerTag)
	log.Print()

	topTags, err := store.TopTags(10)
	if err != nil {
		return fmt.Errorf("could not retrieve top tags: %v", err)
	}

	log.Printf("TOP TAGS")
	log.Print()
	maxLength := 0
	maxCountWidth := 0
	for _, tag := range topTags {
		countWidth := int(math.Log(float64(tag.FileCount)))
		if countWidth > maxCountWidth {
			maxCountWidth = countWidth
		}
		if len(tag.Name) > maxLength {
			maxLength = len(tag.Name)
		}
	}
	for _, tag := range topTags {
		log.Printf("  %*v %*s", maxCountWidth, tag.FileCount, -maxLength, tag.Name)
	}
	log.Print()

	return nil
}