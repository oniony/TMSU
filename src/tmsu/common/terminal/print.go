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

package terminal

import (
	"fmt"
	"sort"
	"strings"
	"tmsu/common/terminal/ansi"
)

const ETX rune = '\003'

func PrintListOnLine(items ansi.Strings, ansi bool) {
	if !ansi {
		items = plain(items)
	}

	sort.Sort(items)

	for index, item := range items {
		if index > 0 {
			fmt.Print(" ")
		}

		fmt.Print(item)
	}

	fmt.Println()
}

func PrintList(items ansi.Strings, indent int, ansi bool) {
	if !ansi {
		items = plain(items)
	}

	sort.Sort(items)

	var padding = strings.Repeat(" ", indent)

	for _, item := range items {
		fmt.Println(padding + string(item))
	}
}

func PrintColumns(items ansi.Strings, width int, ansi bool) {
	if !ansi {
		items = plain(items)
	}

	sort.Sort(items)

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

			var itemLength = item.Length()
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
				padding := (colWidths[columnIndex] + padding) - item.Length()
				fmt.Print(strings.Repeat(" ", padding))
			}
		}

		fmt.Println()
	}
}

func Print(text ansi.String, ansi bool) {
	if !ansi {
		text = text.Plain()
	}

	fmt.Print(string(text))
}

func Println(text ansi.String, ansi bool) {
	if !ansi {
		text = text.Plain()
	}

	fmt.Println(string(text))
}

//TODO doesn't calculate correctly when there are ansi codes
func PrintWrapped(text ansi.String, maxWidth int, ansi bool) {
	if !ansi {
		text = text.Plain()
	}

	if maxWidth == 0 {
		fmt.Println(string(text))
		return
	}

	word := ""
	width := 0
	indent := 0
	for _, r := range string(text) + string(ETX) {
		if width == 0 && word == "" && r == ' ' {
			indent += 1
			continue
		}

		if r == ' ' || r == '\n' || r == ETX {
			charsNeeded := len(word)
			if width > 0 {
				charsNeeded += 1 // space
			}

			if width+charsNeeded > maxWidth {
				fmt.Println()
				width = 0

				if indent > 0 {
					fmt.Print(strings.Repeat(" ", indent))
					width += indent
				}
			} else {
				if width > 0 {
					fmt.Print(" ")
					width += 1
				} else {
					if indent > 0 {
						fmt.Print(strings.Repeat(" ", indent))
						width += indent
					}
				}
			}

			fmt.Print(word)
			width += len(word)
			word = ""

			if r == '\n' {
				fmt.Println()
				width = 0
				indent = 0
			}
		} else {
			word += string(r)
		}
	}

	if width > 0 {
		fmt.Print(" ")
	} else {
		if indent > 0 {
			fmt.Print(strings.Repeat(" ", indent))
		}
	}

	fmt.Println(word)
}

// unexported

func plain(items []ansi.String) []ansi.String {
	plain := make([]ansi.String, len(items))

	for index, item := range items {
		plain[index] = item.Plain()
	}

	return plain
}
