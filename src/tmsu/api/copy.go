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

func CopyTag(store *storage.Storage, tx *storage.Tx, sourceTagName string, destTagName string) error {
	sourceTag, err := store.TagByName(tx, sourceTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
	}
	if sourceTag == nil {
		return NoSuchTag{sourceTagName}
	}

	destTag, err := store.TagByName(tx, destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
	}
	if destTag != nil {
		return TagAlreadyExists{destTagName}
	}

	log.Infof(2, "copying tag '%v' to '%v'.", sourceTagName, destTagName)

	if _, err = store.CopyTag(tx, sourceTag.Id, destTagName); err != nil {
		return fmt.Errorf("could not copy tag '%v' to '%v': %v", sourceTagName, destTagName, err)
	}

	return nil
}
