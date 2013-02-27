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

func TestPaths(test *testing.T) {
	tree := BuildTree("/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m")
	paths := tree.Paths()

	if len(paths) != 7 {
		test.Fatalf("Expected 8 paths but were %v.", len(paths))
	}

	if paths[0] != "/a/b" {
		test.Fatalf("Expected top-level path 0 to be '/a/b' but was '%v'.", paths[0])
	}

	if paths[1] != "/a/b/c" {
		test.Fatalf("Expected top-level path 1 to be '/a/b/c' but was '%v'.", paths[0])
	}

	if paths[2] != "/a/b/d" {
		test.Fatalf("Expected top-level path 2 to be '/a/b/d' but was '%v'.", paths[0])
	}

	if paths[3] != "/a/b/e" {
		test.Fatalf("Expected top-level path 3 to be '/a/b/e' but was '%v'.", paths[0])
	}

	if paths[4] != "/a/f" {
		test.Fatalf("Expected top-level path 4 to be '/a/f' but was '%v'.", paths[0])
	}

	if paths[5] != "/j/k/l" {
		test.Fatalf("Expected top-level path 5 to be '/j/k/l' but was '%v'.", paths[0])
	}

	if paths[6] != "/j/k/m" {
		test.Fatalf("Expected top-level path 6 to be '/j/k/m' but was '%v'.", paths[0])
	}
}

func TestTopLevel(test *testing.T) {
	tree := BuildTree("/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m")
	paths := tree.TopLevel().Paths()

	if len(paths) != 4 {
		test.Fatalf("Expected 4 top-level paths but were %v.", len(paths))
	}
	if paths[0] != "/a/b" {
		test.Fatalf("Expected top-level path 0 to be '/a/b' but was '%v'.", paths[0])
	}
	if paths[1] != "/a/f" {
		test.Fatalf("Expected top-level path 1 to be '/a/f' but was '%v'.", paths[1])
	}
	if paths[2] != "/j/k/l" {
		test.Fatalf("Expected top-level path 2 to be '/j/k/l' but was '%v'.", paths[1])
	}
	if paths[3] != "/j/k/m" {
		test.Fatalf("Expected top-level path 3 to be '/j/k/m' but was '%v'.", paths[1])
	}
}

func TestLeaves(test *testing.T) {
	tree := BuildTree("/a/b/c", "/a/b/d", "/a/b", "/a/b/e", "/a/f", "/a/b", "/j/k/l", "/j/k/m")
	paths := tree.Leaves().Paths()

	if len(paths) != 6 {
		test.Fatalf("Expected 6 leaf paths but were %v.", len(paths))
	}
	if paths[0] != "/a/b/c" {
		test.Fatalf("Expected leaf path 0 to be '/a/b/c' but was '%v'.", paths[0])
	}
	if paths[1] != "/a/b/d" {
		test.Fatalf("Expected leaf path 1 to be '/a/b/d' but was '%v'.", paths[1])
	}
	if paths[2] != "/a/b/e" {
		test.Fatalf("Expected leaf path 2 to be '/a/b/e' but was '%v'.", paths[1])
	}
	if paths[3] != "/a/f" {
		test.Fatalf("Expected leaf path 3 to be '/a/f' but was '%v'.", paths[1])
	}
	if paths[4] != "/j/k/l" {
		test.Fatalf("Expected leaf path 3 to be '/j/k/l' but was '%v'.", paths[1])
	}
	if paths[5] != "/j/k/m" {
		test.Fatalf("Expected leaf path 3 to be '/j/k/m' but was '%v'.", paths[1])
	}
}
