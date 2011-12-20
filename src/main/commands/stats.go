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
package main
*/

package main

import (
    "fmt"
)

type StatsCommand struct {}

func (this StatsCommand) Name() string {
    return "stats"
}

func (this StatsCommand) Summary() string {
    return "shows database statistics"
}

func (this StatsCommand) Help() string {
    return `tmsu stats
tmsu stats

Shows the database statistics.`
}

func (this StatsCommand) Exec(args []string) error {
    db, error := OpenDatabase(databasePath())
    if error != nil { return error }
    defer db.Close()

    tagCount, error := db.TagCount()
    if error != nil { return error }

    fileCount, error := db.FileCount()
    if error != nil { return error }

    fileTagCount, error := db.FileTagCount()
    if error != nil { return error }

    fmt.Printf("Database Contents\n")

    fmt.Printf(" Tags:             %v\n", tagCount)
    fmt.Printf(" Files:            %v\n", fileCount)
    fmt.Printf(" Taggings:         %v\n", fileTagCount)
    fmt.Printf("   Avg. per file:  %v\n", fileTagCount / fileCount) 

    return nil
}
