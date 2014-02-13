/*

Copyright 2011-2014 Paul Ruane.

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
)

func columns(items []string, width int) {
	padding := 2 // minimum column padding

	// calculate number of columns to start with
	cols := width / 5
	if cols < 1 {
		cols = 1
	}

	var rows int
	var colWidths []int
	var totalWidth int

	// reduce number of columns until everything fits
	for totalWidth = width + 1; totalWidth > width && cols > 0; cols-- {
		// calculate number of rows for this many columns
		rows = len(items) / cols
		if len(items)%cols != 0 {
			rows++
		}

		// calculated number of columns for this many rows
		// as a row increase above may have changed the picture
		cols = len(items) / rows
		if len(items)%rows > 0 {
			cols++
		}

		// calculate column widths and total width
		colWidths = make([]int, cols)
		totalWidth = cols * padding
		for index, item := range items {
			columnIndex := index / rows

			if len(item) > colWidths[columnIndex] {
				totalWidth += -colWidths[columnIndex] + len(item)
				colWidths[columnIndex] = len(item)
			}
		}
	}
	cols++

	// apportion any remaining space between the columns
	if cols > 2 && rows > 1 {
		padding = (width-totalWidth)/(cols-1) + 2
		if padding < 2 {
			padding = 2
		}
	}

	for rowIndex := 0; rowIndex < rows; rowIndex += 1 {
		for columnIndex := 0; columnIndex < cols; columnIndex++ {
			itemIndex := rows*columnIndex + rowIndex

			if itemIndex >= len(items) {
				break
			}

			item := items[itemIndex]

			fmt.Print(item)

			if columnIndex < cols-1 {
				padding := (colWidths[columnIndex] + padding) - len(item)
				fmt.Print(strings.Repeat(" ", padding))
			}
		}

		fmt.Println()
	}
}
