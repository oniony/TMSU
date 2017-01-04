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

package terminal

import (
	"fmt"
	"github.com/oniony/TMSU/common/terminal/ansi"
	"strings"
)

const ETX rune = '\003'

func PrintColumns(items []string) {
	PrintColumnsWidth(items, Width())
}

func PrintColumnsWidth(items []string, width int) {
	ansi.Sort(items)

	padding := 2 // minimum column padding

	var colWidths []int
	var calcWidth int

	cols := 0
	rows := 1

	// add a row until everything fits or we have every item on its own row
	for calcWidth = width + 1; calcWidth > width && rows <= len(items); rows++ {
		cols = 0
		colWidths = make([]int, 0, width)
		calcWidth = -padding // last column has no padding

		// try to place items into columns
		for index, item := range items {
			col := index / rows

			if col >= len(colWidths) {
				// add column
				cols++
				colWidths = append(colWidths, 0)
				calcWidth += padding
			}

			itemLength := len(ansi.Strip(item))
			if itemLength > colWidths[col] {
				// widen column
				calcWidth += -colWidths[col] + itemLength
				colWidths[col] = itemLength
			}

			if calcWidth > width {
				// exceeded width
				break
			}
		}
	}
	rows--

	// apportion any remaining space between the columns
	if cols > 2 && rows > 1 {
		padding = (width-calcWidth)/(cols-1) + 2
		if padding < 2 {
			padding = 2
		}
	}

	// render
	for rowIndex := 0; rowIndex < rows; rowIndex++ {
		for columnIndex := 0; columnIndex < cols; columnIndex++ {
			itemIndex := rows*columnIndex + rowIndex

			if itemIndex >= len(items) {
				break
			}

			item := items[itemIndex]

			fmt.Print(item)

			if columnIndex < cols-1 {
				itemLength := len(ansi.Strip(item))
				padding := (colWidths[columnIndex] + padding) - itemLength
				fmt.Print(strings.Repeat(" ", padding))
			}
		}

		fmt.Println()
	}
}

func PrintWrapped(text string) {
	PrintWrappedWidth(text, Width())
}

func PrintWrappedWidth(text string, maxWidth int) {
	if maxWidth == 0 {
		fmt.Println(string(text))
		return
	}

	word := ""
	width := 0
	indent := 0
	for _, r := range string(text) + string(ETX) {
		if r == ' ' || r == '\n' || r == ETX {
			// tabulation
			if ansi.Strip(word) == "" && r == ' ' {
				fmt.Print(" ")
				width += 1
				indent = width
				continue
			}

			charsNeeded := len(word)
			if width > 0 {
				charsNeeded += 1 // space
			}

			if width+charsNeeded > maxWidth {
				// wrap onto new line
				fmt.Println()
				width = 0

				if indent > 0 {
					// print indent on new line
					fmt.Print(strings.Repeat(" ", indent))
					width += indent
				}
			} else {
				if width > indent {
					// add space between words
					fmt.Print(" ")
					width += 1
				}
			}

			if ansi.Strip(word) == "" && indent > 0 {
				// print indent on new line
				fmt.Print(strings.Repeat(" ", indent))
				width += indent
			}

			fmt.Print(word)
			width += len(ansi.Strip(word))
			word = ""

			if r == '\n' {
				// start a new line
				fmt.Println()
				width = 0
				indent = 0
			}
		} else {
			// add character to word
			word += string(r)
		}
	}

	fmt.Println()
}
