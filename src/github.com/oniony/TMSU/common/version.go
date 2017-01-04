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

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major uint
	Minor uint
	Patch uint
}

func ParseVersion(version string) Version {
	parts := strings.Split(version, ".")

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		panic("invalid major version")
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		panic("invalid minor version")
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		panic("invalid patch version")
	}

	return Version{uint(major), uint(minor), uint(patch)}
}

func (version Version) String() string {
	return fmt.Sprintf("%v.%v.%v", version.Major, version.Minor, version.Patch)
}

func (this Version) LessThan(that Version) bool {
	return this.Major < that.Major ||
		(this.Major == that.Major && this.Minor < that.Minor) ||
		(this.Major == that.Major && this.Minor == that.Minor && this.Patch < that.Patch)
}

func (this Version) GreaterThan(that Version) bool {
	return this.Major > that.Major ||
		(this.Major == that.Major && this.Minor > that.Minor) ||
		(this.Major == that.Major && this.Minor == that.Minor && this.Patch > that.Patch)
}
