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
	"github.com/oniony/TMSU/storage/database"
)

// Retrieves the complete set of tag implications.
func (storage *Storage) Implications(tx *Tx) (entities.Implications, error) {
	return database.Implications(tx.tx)
}

// Retrieves the set of implications for the specified tag and value pairs.
func (storage *Storage) ImplicationsFor(tx *Tx, pairs ...entities.TagIdValueIdPair) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	impliedPairs := make(entities.TagIdValueIdPairs, len(pairs))
	copy(impliedPairs, pairs)

	for len(impliedPairs) > 0 {
		implications, err := database.ImplicationsFor(tx.tx, impliedPairs)
		if err != nil {
			return nil, err
		}

		impliedPairs = make(entities.TagIdValueIdPairs, 0)
		for _, implication := range implications {
			if !resultantImplications.Contains(*implication) {
				resultantImplications = append(resultantImplications, implication)
				impliedPairs = append(impliedPairs, entities.TagIdValueIdPair{implication.ImpliedTag.Id, implication.ImpliedValue.Id})
			}
		}
	}

	return resultantImplications, nil
}

// Retrieves the set of implications that imply the specified tag and value pairs.
func (storage *Storage) ImplicationsImplying(tx *Tx, pairs ...entities.TagIdValueIdPair) (entities.Implications, error) {
	resultantImplications := make(entities.Implications, 0)

	implyingPairs := make(entities.TagIdValueIdPairs, len(pairs))
	copy(implyingPairs, pairs)

	for len(implyingPairs) > 0 {
		implications, err := database.ImplyingImplications(tx.tx, implyingPairs)
		if err != nil {
			return nil, err
		}

		implyingPairs = make(entities.TagIdValueIdPairs, 0)
		for _, implication := range implications {
			if resultantImplications.Contains(*implication) {
				resultantImplications = append(resultantImplications, implication)
				implyingPairs = append(implyingPairs, entities.TagIdValueIdPair{implication.ImplyingTag.Id, implication.ImplyingValue.Id})
			}
		}
	}

	return resultantImplications, nil
}

// Adds the specified implication.
func (storage Storage) AddImplication(tx *Tx, pair, impliedPair entities.TagIdValueIdPair) error {
	implications, err := storage.ImplicationsFor(tx, impliedPair)
	if err != nil {
		return err
	}

	for _, implication := range implications {
		if implication.ImpliedTag.Id == pair.TagId && (pair.ValueId == 0 || implication.ImpliedValue.Id == pair.ValueId) {
			return fmt.Errorf("implication would create a cycle")
		}
	}

	return database.AddImplication(tx.tx, pair, impliedPair)
}

// Deletes the specified implication
func (storage Storage) DeleteImplication(tx *Tx, pair, impliedPair entities.TagIdValueIdPair) error {
	return database.DeleteImplication(tx.tx, pair, impliedPair)
}

// Deletes implications for the specified tag.
func (storage Storage) DeleteImplicationsByTagId(tx *Tx, tagId entities.TagId) error {
	return database.DeleteImplicationsByTagId(tx.tx, tagId)
}

// Deletes implications for the specified value.
func (storage Storage) DeleteImplicationsByValueId(tx *Tx, valueId entities.ValueId) error {
	return database.DeleteImplicationsByValueId(tx.tx, valueId)
}
