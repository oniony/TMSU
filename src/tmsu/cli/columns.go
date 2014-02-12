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

	cols := width / 5 // assume column width is at least five characters
	if cols < 1 {
		cols = 1
	}

	var rows int
	var colWidths []int
	var totalWidth int
	for totalWidth = width + 1; totalWidth > width && cols > 0; cols-- {
		colWidths = make([]int, cols)

		rows = len(items) / cols

		// add a row if necessary
		if len(items)%cols != 0 {
			rows++
		}

		minimumColumns := len(items) / rows
		if len(items)%rows > 0 {
			minimumColumns++
		}

		if minimumColumns < cols {
			cols = minimumColumns + 1
			continue
		}

		for index, item := range items {
			columnIndex := index / rows

			if len(item) > colWidths[columnIndex] {
				colWidths[columnIndex] = len(item)
			}
		}

		totalWidth = 0
		for _, colWidth := range colWidths {
			totalWidth += colWidth
			totalWidth += padding
		}
		totalWidth -= padding
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
