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

package ansi

import (
	"regexp"
)

type String string
type Strings []String

func (items Strings) Len() int {
	return len(items)
}

func (items Strings) Less(i, j int) bool {
	return Plain(items[i]) < Plain(items[j])
}

func (items Strings) Swap(i, j int) {
	items[j], items[i] = items[i], items[j]
}

func Length(item String) int {
	return len(formatting.ReplaceAllLiteralString(string(item), ""))
}

func Sort(items Strings) {
	//TODO
}

func Plain(item String) string {
	return formatting.ReplaceAllLiteralString(string(item), "")
}

// unexported

var formatting = regexp.MustCompile(`\x1b\[[0-9]*(;[0-9]*)*m`)
