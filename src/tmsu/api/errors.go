// Copyright 2011-2015 Paul Ruane.

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

package api

import (
	"fmt"
)

type NoSuchTag struct {
	TagName string
}

func (err NoSuchTag) Error() string {
	return fmt.Sprintf("no such tag '%v'.", err.TagName)
}

type TagAlreadyExists struct {
	TagName string
}

func (err TagAlreadyExists) Error() string {
	return fmt.Sprintf("tag '%v' already exists.", err.TagName)
}
