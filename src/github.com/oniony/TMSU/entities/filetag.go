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

type FileTag struct {
	FileId   FileId
	TagId    TagId
	ValueId  ValueId
	Explicit bool
	Implicit bool
}

func (fileTag FileTag) ToTagIdValueIdPair() TagIdValueIdPair {
	return TagIdValueIdPair{fileTag.TagId, fileTag.ValueId}
}

type FileTags []*FileTag

func (fileTags FileTags) ToTagIdValueIdPairs() TagIdValueIdPairs {
	pairs := make(TagIdValueIdPairs, len(fileTags))

	for index, fileTag := range fileTags {
		pairs[index] = fileTag.ToTagIdValueIdPair()
	}

	return pairs
}

func (fileTags FileTags) Any(predicate func(fileTag FileTag) bool) bool {
	for _, fileTag := range fileTags {
		if predicate(*fileTag) {
			return true
		}
	}

	return false
}

func (fileTags FileTags) Where(predicate func(fileTag FileTag) bool) FileTags {
	matches := make(FileTags, 0, 10)

	for _, fileTag := range fileTags {
		if predicate(*fileTag) {
			matches = append(matches, fileTag)
		}
	}

	return matches
}

func (fileTags FileTags) Single() *FileTag {
	switch len(fileTags) {
	case 1:
		return fileTags[0]
	default:
		return nil
	}
}

func (fileTags FileTags) FileIds() FileIds {
	fileIds := make(FileIds, len(fileTags))
	for index, fileTag := range fileTags {
		fileIds[index] = fileTag.FileId
	}

	return fileIds.Uniq()
}

func (fileTags FileTags) TagIds() TagIds {
	tagIds := make(TagIds, len(fileTags))
	for index, fileTag := range fileTags {
		tagIds[index] = fileTag.TagId
	}

	return tagIds.Uniq()
}

func (fileTags FileTags) ValueIds() ValueIds {
	valueIds := make(ValueIds, len(fileTags))
	for index, fileTag := range fileTags {
		valueIds[index] = fileTag.ValueId
	}

	return valueIds.Uniq()
}
