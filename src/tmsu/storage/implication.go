/*
Copyright 2011-2013 Paul Ruane.

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

package storage

import (
	"tmsu/storage/database"
)

// Retrieves the complete set of tag implications.
func (storage *Storage) Implications() (database.Implications, error) {
	return storage.Db.Implications()
}

// Retrieves the set of implications for the specified tags.
func (storage *Storage) ImplicationsForTags(tags database.Tags) (database.Implications, error) {
	implications, err := storage.Db.ImplicationsForTags(tags)
	if err != nil {
		return nil, err
	}

	if len(implications) > 0 {
		tags = make(database.Tags, len(implications))
		for index, implication := range implications {
			tags[index] = &implication.ImpliedTag
		}

		recursiveImplications, err := storage.ImplicationsForTags(tags)
		if err != nil {
			return nil, err
		}

		implications = append(implications, recursiveImplications...)
	}

	return implications, nil
}

// Adds the specified implication.
func (storage Storage) AddImplication(tagId, impliedTagId uint) error {
	return storage.Db.AddImplication(tagId, impliedTagId)
}

// Updates implications featuring the specified tag.
func (storage Storage) UpdateImplicationsForTagId(tagId, impliedTagId uint) error {
	return storage.Db.UpdateImplicationsForTagId(tagId, impliedTagId)
}

// Removes the specified implication
func (storage Storage) RemoveImplication(tagId, impliedTagId uint) error {
	return storage.Db.DeleteImplication(tagId, impliedTagId)
}

// Removes implications featuring the specified tag.
func (storage Storage) RemoveImplicationsForTagId(tagId uint) error {
	return storage.Db.DeleteImplicationsForTagId(tagId)
}
