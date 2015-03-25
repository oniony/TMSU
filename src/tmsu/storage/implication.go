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

// Retrieves the set of implications for the specified tags.
func (storage *Storage) ImplicationsForTags(tx *Tx, tagIds ...entities.TagId) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	impliedTagIds := make(entities.TagIds, len(tagIds))
	copy(impliedTagIds, tagIds)

	for len(impliedTagIds) > 0 {
		implications, err := database.ImplicationsForTags(tx.tx, impliedTagIds)
		if err != nil {
			return nil, err
		}

		impliedTagIds = make(entities.TagIds, 0)
		for _, implication := range implications {
			if !containsImplication(resultantImplications, implication) {
				resultantImplications = append(resultantImplications, implication)
				impliedTagIds = append(impliedTagIds, implication.ImpliedTag.Id)
			}
		}
	}

	return resultantImplications, nil
}

// Adds the specified implication.
func (storage Storage) AddImplication(tx *Tx, tagId, impliedTagId entities.TagId) error {
	return database.AddImplication(tx.tx, tagId, impliedTagId)
}

// Updates implications featuring the specified tag.
func (storage Storage) UpdateImplicationsForTagId(tx *Tx, tagId, impliedTagId entities.TagId) error {
	return database.UpdateImplicationsForTagId(tx.tx, tagId, impliedTagId)
}

// Removes the specified implication
func (storage Storage) RemoveImplication(tx *Tx, tagId, impliedTagId entities.TagId) error {
	return database.DeleteImplication(tx.tx, tagId, impliedTagId)
}

// Removes implications featuring the specified tag.
func (storage Storage) RemoveImplicationsForTagId(tx *Tx, tagId entities.TagId) error {
	return database.DeleteImplicationsForTagId(tx.tx, tagId)
}

// unexported

func containsImplication(implications entities.Implications, implication *entities.Implication) bool {
	for index := 0; index < len(implications); index++ {
		if implications[index].ImplyingTag.Id == implication.ImplyingTag.Id && implications[index].ImpliedTag.Id == implication.ImpliedTag.Id {
			return true
		}
	}

	return false
}
