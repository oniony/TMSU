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

package api

import (
	"fmt"
	"tmsu/common/log"
	"tmsu/storage"
)

func DeleteTag(store *storage.Storage, tx *storage.Tx, tagName string) error {
	tag, err := store.TagByName(tx, tagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
	}
	if tag == nil {
		return NoSuchTag{tagName}
	}

	log.Infof(2, "deleting tag '%v'.", tagName)

	if err = store.DeleteTag(tx, tag.Id); err != nil {
		return fmt.Errorf("could not delete tag '%v': %v", tagName, err)
	}

	return nil
}
