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

package entities

type Implication struct {
	ImplyingTag   Tag
	ImplyingValue Value
	ImpliedTag    Tag
	ImpliedValue  Value
}

func (implication Implication) ImplyingTagValuePair() TagValuePair {
	return TagValuePair{implication.ImplyingTag.Id, implication.ImplyingValue.Id}
}

func (implication Implication) ImpliedTagValuePair() TagValuePair {
	return TagValuePair{implication.ImpliedTag.Id, implication.ImpliedValue.Id}
}

type Implications []*Implication

func (implications Implications) Any(predicate func(Implication) bool) bool {
	for _, implication := range implications {
		if predicate(*implication) {
			return true
		}
	}

	return false
}

func (implications Implications) Where(predicate func(Implication) bool) Implications {
	result := make(Implications, 0, 10)

	for _, implication := range implications {
		if predicate(*implication) {
			result = append(result, implication)
		}
	}

	return result
}

func (implications Implications) Implies(tagValuePair TagValuePair) bool {
	for _, implication := range implications {
		if implication.ImpliedTag.Id == tagValuePair.TagId && implication.ImpliedValue.Id == tagValuePair.ValueId {
			return true
		}
	}

	return false
}
