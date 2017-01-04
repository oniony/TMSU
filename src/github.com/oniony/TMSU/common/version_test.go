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

package common

import "testing"

func TestLessThan(t *testing.T) {
	v1 := ParseVersion("0.0.0")
	v2 := ParseVersion("0.0.1")
	v3 := ParseVersion("0.1.0")
	v4 := ParseVersion("1.0.0")

	if !v1.LessThan(v2) || v2.LessThan(v1) {
		t.Fatalf("%v should be less than %v", v1.String(), v2.String())
	}

	if !v2.LessThan(v3) || v3.LessThan(v2) {
		t.Fatalf("%v should be less than %v", v2.String(), v3.String())
	}

	if !v3.LessThan(v4) || v4.LessThan(v3) {
		t.Fatalf("%v should be less than %v", v3.String(), v4.String())
	}
}

func TestGreaterThan(t *testing.T) {
	v1 := ParseVersion("1.0.0")
	v2 := ParseVersion("0.1.0")
	v3 := ParseVersion("0.0.1")
	v4 := ParseVersion("0.0.0")

	if !v1.GreaterThan(v2) || v2.GreaterThan(v1) {
		t.Fatalf("%v should be greater than %v", v1.String(), v2.String())
	}

	if !v2.GreaterThan(v3) || v3.GreaterThan(v2) {
		t.Fatalf("%v should be greater than %v", v2.String(), v3.String())
	}

	if !v3.GreaterThan(v4) || v4.GreaterThan(v3) {
		t.Fatalf("%v should be greater than %v", v3.String(), v4.String())
	}
}
