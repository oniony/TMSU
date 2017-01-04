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

package database

import (
	"fmt"
	"github.com/oniony/TMSU/common"
)

// unexported

type schemaVersion struct {
	common.Version
	Revision uint
}

func (version schemaVersion) String() string {
	return fmt.Sprintf("%v.%v.%v-%v", version.Major, version.Minor, version.Patch, version.Revision)
}

func (this schemaVersion) LessThan(that schemaVersion) bool {
	return this.Major < that.Major || this.Minor < that.Minor || this.Patch < that.Patch || this.Revision < that.Revision
}

func (this schemaVersion) GreaterThan(that schemaVersion) bool {
	return this.Major > that.Major || this.Minor > that.Minor || this.Patch > that.Patch || this.Revision > that.Revision
}
