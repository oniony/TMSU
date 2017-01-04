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

type Implication struct {
	ImplyingTag   Tag
	ImplyingValue Value
	ImpliedTag    Tag
	ImpliedValue  Value
}

func (implication Implication) ImplyingTagValuePair() TagIdValueIdPair {
	return TagIdValueIdPair{implication.ImplyingTag.Id, implication.ImplyingValue.Id}
}

func (implication Implication) ImpliedTagValuePair() TagIdValueIdPair {
	return TagIdValueIdPair{implication.ImpliedTag.Id, implication.ImpliedValue.Id}
}

type Implications []*Implication

func (implications Implications) Contains(implication Implication) bool {
	for _, i := range implications {
		if i.ImplyingTag.Id == implication.ImplyingTag.Id && i.ImplyingValue.Id == implication.ImplyingValue.Id &&
			i.ImpliedTag.Id == implication.ImpliedTag.Id && i.ImpliedValue.Id == implication.ImpliedValue.Id {
			return true
		}
	}

	return false
}

func (implications Implications) Any(predicate func(Implication) bool) bool {
	for _, implication := range implications {
		if predicate(*implication) {
			return true
		}
	}

	return false
}

func (implications Implications) Where(predicate func(Implication) bool) Implications {
	matches := make(Implications, 0, 10)

	for _, implication := range implications {
		if predicate(*implication) {
			matches = append(matches, implication)
		}
	}

	return matches
}

func (implications Implications) Implies(tagValuePair TagIdValueIdPair) bool {
	for _, implication := range implications {
		if implication.ImpliedTag.Id == tagValuePair.TagId && implication.ImpliedValue.Id == tagValuePair.ValueId {
			return true
		}
	}

	return false
}
