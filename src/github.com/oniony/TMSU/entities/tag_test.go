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

package entities

import (
	"testing"
)

func TestUniqueTagIds(test *testing.T) {
	// set-up

	tagIds := TagIds{1, 2, 3, 2, 1, 4, 4, 1}

	// test

	uniq := tagIds.Uniq()

	// validate

	if len(uniq) != 4 || uniq[0] != TagId(1) || uniq[1] != TagId(2) || uniq[2] != TagId(3) || uniq[3] != TagId(4) {
		test.Fatalf("Unexpected unique set: %v", uniq)
	}
}
