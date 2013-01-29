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

package path

import (
	"testing"
)

func TestRoots(test *testing.T) {
	roots, err := Roots([]string{"/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m"})
	if err != nil {
		test.Fatalf("Unexpected error: %v", err)
	}

	if len(roots) != 4 {
		test.Fatalf("Expected 4 root paths but were %v.", len(roots))
	}
	if roots[0] != "/a/b" {
		test.Fatalf("Expected root path 0 to be '/a/b' but was '%v'.", roots[0])
	}
	if roots[1] != "/a/f" {
		test.Fatalf("Expected root path 1 to be '/a/f' but was '%v'.", roots[1])
	}
	if roots[2] != "/j/k/l" {
		test.Fatalf("Expected root path 2 to be '/j/k/l' but was '%v'.", roots[1])
	}
	if roots[3] != "/j/k/m" {
		test.Fatalf("Expected root path 3 to be '/j/k/m' but was '%v'.", roots[1])
	}
}
