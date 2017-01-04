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

package storage

import (
	"fmt"
	"github.com/oniony/TMSU/entities"
)

type AbsolutePathResolutionError struct {
	Path   string
	Reason error
}

func (err AbsolutePathResolutionError) Error() string {
	return fmt.Sprintf("Cannot resolve absolute path '%v': %v", err.Path, err.Reason)
}

type FileTagDoesNotExist struct {
	FileId  entities.FileId
	TagId   entities.TagId
	ValueId entities.ValueId
}

func (err FileTagDoesNotExist) Error() string {
	return fmt.Sprintf("File-tag for file #%v, tag #%v and value #%v does not exist", err.FileId, err.TagId, err.ValueId)
}
