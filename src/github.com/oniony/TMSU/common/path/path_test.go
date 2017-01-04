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

package path

import (
	"testing"
)

func TestRelTo(test *testing.T) {
	paths := map[string]string{
		"/":                      "./some/path",
		"/other":                 "../some/path",
		"/other/":                "../some/path",
		"/other/mother":          "/some/path",
		"/other/mother/":         "/some/path",
		"/some":                  "./path",
		"/some/":                 "./path",
		"/some/path":             ".",
		"/some/path/":            ".",
		"/some/cheese":           "../path",
		"/some/cheese/":          "../path",
		"/some/cheese/sandwich":  "/some/path",
		"/some/cheese/sandwich/": "/some/path"}

	for to, expected := range paths {
		actual := RelTo("/some/path", to)

		if actual != expected {
			test.Fatalf("Expected '/some/path' relative to '%v' to be '%v' but was '%v'", to, expected, actual)
		}
	}
}
