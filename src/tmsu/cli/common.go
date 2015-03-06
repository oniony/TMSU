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

package cli

import (
	"errors"
	"fmt"
	"os"
	"time"
	"tmsu/common/log"
	"tmsu/common/terminal"
	"tmsu/entities"
	"tmsu/storage"
)

// unexported

var errBlank = errors.New("")

type tagValuePair struct {
	TagId   entities.TagId
	ValueId entities.ValueId
}

func useColour(options Options) (bool, error) {
	when := "auto"
	if options.HasOption("--color") {
		when = options.Get("--color").Argument
	}

	switch when {
	case "":
	case "auto":
		return terminal.Colour() && terminal.Width() > 0, nil
	case "always":
		return true, nil
	case "never":
		return false, nil
	}

	return false, fmt.Errorf("invalid argument '%v' for '--color'", when)
}

type emptyStat struct {
	name string
}

func (es emptyStat) Name() string {
	return es.name
}

func (emptyStat) Size() int64 {
	return 0
}

func (emptyStat) Mode() os.FileMode {
	return 0
}

func (emptyStat) ModTime() time.Time {
	return time.Time{}
}

func (emptyStat) IsDir() bool {
	return false
}

func (emptyStat) Sys() interface{} {
	return nil
}

func createTag(store *storage.Storage, tagName string) (*entities.Tag, error) {
	tag, err := store.AddTag(tagName)
	if err != nil {
		return nil, fmt.Errorf("could not create tag '%v': %v", tagName, err)
	}

	log.Warnf("New tag '%v'.", tagName)

	return tag, nil
}

func createValue(store *storage.Storage, valueName string) (*entities.Value, error) {
	value, err := store.AddValue(valueName)
	if err != nil {
		return nil, err
	}

	log.Warnf("New value '%v'.", valueName)

	return value, nil
}
