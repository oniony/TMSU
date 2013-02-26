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

func TestTopLevel(test *testing.T) {
	topLevel, err := TopLevel([]string{"/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m"})
	if err != nil {
		test.Fatalf("Unexpected error: %v", err)
	}

	if len(topLevel) != 4 {
		test.Fatalf("Expected 4 top-level paths but were %v.", len(topLevel))
	}
	if topLevel[0] != "/a/b" {
		test.Fatalf("Expected top-level path 0 to be '/a/b' but was '%v'.", topLevel[0])
	}
	if topLevel[1] != "/a/f" {
		test.Fatalf("Expected top-level path 1 to be '/a/f' but was '%v'.", topLevel[1])
	}
	if topLevel[2] != "/j/k/l" {
		test.Fatalf("Expected top-level path 2 to be '/j/k/l' but was '%v'.", topLevel[1])
	}
	if topLevel[3] != "/j/k/m" {
		test.Fatalf("Expected top-level path 3 to be '/j/k/m' but was '%v'.", topLevel[1])
	}
}

func TestLeaves(test *testing.T) {
	leaves, err := Leaves([]string{"/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m"})
	if err != nil {
		test.Fatalf("Unexpected error: %v", err)
	}

	if len(leaves) != 6 {
		test.Fatalf("Expected 4 top-level paths but were %v.", len(leaves))
	}
	if leaves[0] != "/a/b/c" {
		test.Fatalf("Expected leaf path 0 to be '/a/b/c' but was '%v'.", leaves[0])
	}
	if leaves[1] != "/a/b/d" {
		test.Fatalf("Expected leaf path 1 to be '/a/b/d' but was '%v'.", leaves[1])
	}
	if leaves[2] != "/a/b/e" {
		test.Fatalf("Expected leaf path 2 to be '/a/b/e' but was '%v'.", leaves[1])
	}
	if leaves[3] != "/a/f" {
		test.Fatalf("Expected leaf path 3 to be '/a/f' but was '%v'.", leaves[1])
	}
	if leaves[4] != "/j/k/l" {
		test.Fatalf("Expected leaf path 3 to be '/j/k/l' but was '%v'.", leaves[1])
	}
	if leaves[5] != "/j/k/m" {
		test.Fatalf("Expected leaf path 3 to be '/j/k/m' but was '%v'.", leaves[1])
	}
}
