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

package storage

import (
	"tmsu/entities"
	"tmsu/storage/database"
)

// Retrieves the complete set of tag implications.
func (storage *Storage) Implications(tx *Tx) (entities.Implications, error) {
	return database.Implications(tx.tx)
}

// Retrieves the set of implications for the specified tag and value pairs.
func (storage *Storage) ImplicationsFor(tx *Tx, tagValuePairs ...entities.TagValuePair) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	impliedTagValuePairs := make(entities.TagValuePairs, len(tagValuePairs))
	copy(impliedTagValuePairs, tagValuePairs)

	for len(impliedTagValuePairs) > 0 {
		implications, err := database.ImplicationsFor(tx.tx, impliedTagValuePairs)
		if err != nil {
			return nil, err
		}

		impliedTagValuePairs = make(entities.TagValuePairs, 0)
		for _, implication := range implications {
			if !containsImplication(resultantImplications, implication) {
				resultantImplications = append(resultantImplications, implication)
				impliedTagValuePairs = append(impliedTagValuePairs, entities.TagValuePair{implication.ImpliedTag.Id, implication.ImpliedValue.Id})
			}
		}
	}

	return resultantImplications, nil
}

// Retrieves the set of implications that imply the specified tag and value pairs.
func (storage *Storage) ImplicationsImplying(tx *Tx, tagValuePairs ...entities.TagValuePair) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	implyingTagValuePairs := make(entities.TagValuePairs, len(tagValuePairs))
	copy(implyingTagValuePairs, tagValuePairs)

	for len(implyingTagValuePairs) > 0 {
		implications, err := database.ImplyingImplications(tx.tx, implyingTagValuePairs)
		if err != nil {
			return nil, err
		}

		implyingTagValuePairs = make(entities.TagValuePairs, 0)
		for _, implication := range implications {
			if !containsImplication(resultantImplications, implication) {
				resultantImplications = append(resultantImplications, implication)
				implyingTagValuePairs = append(implyingTagValuePairs, entities.TagValuePair{implication.ImplyingTag.Id, implication.ImplyingValue.Id})
			}
		}
	}

	return resultantImplications, nil
}

// Adds the specified implication.
func (storage Storage) AddImplication(tx *Tx, tagValuePair, impliedTagValuePair entities.TagValuePair) error {
	return database.AddImplication(tx.tx, tagValuePair, impliedTagValuePair)
}

// Deletes the specified implication
func (storage Storage) DeleteImplication(tx *Tx, tagValuePair, impliedTagValuePair entities.TagValuePair) error {
	return database.DeleteImplication(tx.tx, tagValuePair, impliedTagValuePair)
}

// Deletes implications for the specified tag.
func (storage Storage) DeleteImplicationsByTagId(tx *Tx, tagId entities.TagId) error {
	return database.DeleteImplicationsByTagId(tx.tx, tagId)
}

// Deletes implications for the specified value.
func (storage Storage) DeleteImplicationsByValueId(tx *Tx, valueId entities.ValueId) error {
	return database.DeleteImplicationsByValueId(tx.tx, valueId)
}

// unexported

func containsImplication(implications entities.Implications, check *entities.Implication) bool {
	for _, implication := range implications {
		if implication.ImplyingTag.Id == check.ImplyingTag.Id &&
			implication.ImplyingValue.Id == check.ImplyingValue.Id &&
			implication.ImpliedTag.Id == check.ImpliedTag.Id &&
			implication.ImpliedValue.Id == check.ImpliedValue.Id {
			return true
		}
	}

	return false
}
