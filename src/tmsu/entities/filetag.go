/*
Copyright 2011-2014 Paul Ruane.

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

package entities

type FileTag struct {
	FileId   uint
	TagId    uint
	ValueId  uint
	Explicit bool
	Implicit bool
}

type FileTags []*FileTag

func (fileTags FileTags) Contains(tagId, valueId uint) bool {
	for _, fileTag := range fileTags {
		if fileTag.TagId == tagId && fileTag.ValueId == valueId {
			return true
		}
	}

	return false
}
