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

package ansi

import (
	"regexp"
	"sort"
)

func Bold(text string) string {
	return BoldCode + text + ResetCode
}

func Italic(text string) string {
	return ItalicCode + text + ResetCode
}

func Underline(text string) string {
	return UnderlineCode + text + ResetCode
}

func Blink(text string) string {
	return BlinkCode + text + ResetCode
}

func Invert(text string) string {
	return InvertCode + text + ResetCode
}

func Black(text string) string {
	return BlackCode + text + ResetCode
}

func Red(text string) string {
	return RedCode + text + ResetCode
}

func Green(text string) string {
	return GreenCode + text + ResetCode
}

func Yellow(text string) string {
	return YellowCode + text + ResetCode
}

func Blue(text string) string {
	return BlueCode + text + ResetCode
}

func Magenta(text string) string {
	return MagentaCode + text + ResetCode
}

func Cyan(text string) string {
	return CyanCode + text + ResetCode
}

func White(text string) string {
	return WhiteCode + text + ResetCode
}

func DarkGrey(text string) string {
	return DarkGreyCode + text + ResetCode
}

func Strip(text string) string {
	return formatting.ReplaceAllLiteralString(string(text), "")
}

func Sort(items []string) {
	sort.Sort(ansiStrings(items))
}

// unexported

var formatting = regexp.MustCompile(`\x1b\[[0-9]*(;[0-9]*)*m`)

type ansiString string
type ansiStrings []string

func (items ansiStrings) Len() int {
	return len(items)
}

func (items ansiStrings) Less(i, j int) bool {
	return Strip(items[i]) < Strip(items[j])
}

func (items ansiStrings) Swap(i, j int) {
	items[j], items[i] = items[i], items[j]
}

func (item ansiString) Length() int {
	return len(Strip(string(item)))
}
