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

package cli

import (
	"bytes"
	"fmt"
	"github.com/oniony/TMSU/common/log"
	"github.com/oniony/TMSU/common/terminal"
	"github.com/oniony/TMSU/common/terminal/ansi"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage"
	"github.com/oniony/TMSU/storage/database"
	"os"
	"strings"
	"time"
)

// unexported

func openDatabase(path string) (*storage.Storage, error) {
	storage, err := storage.OpenAt(path)
	if err != nil {
		switch err.(type) {
		case database.DatabaseNotFoundError:
			return nil, fmt.Errorf("no database found: use 'tmsu init' to create one")
		case database.DatabaseAccessError:
			return nil, fmt.Errorf("cannot access database: %v", err)
		default:
			return nil, err
		}
	}

	return storage, nil
}

func stdoutIsCharDevice() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return true
	}

	return false
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

func createTag(store *storage.Storage, tx *storage.Tx, tagName string) (*entities.Tag, error) {
	tag, err := store.AddTag(tx, tagName)
	if err != nil {
		return nil, err
	}

	log.Warnf("new tag '%v'", tagName)

	return tag, nil
}

func createValue(store *storage.Storage, tx *storage.Tx, valueName string) (*entities.Value, error) {
	value, err := store.AddValue(tx, valueName)
	if err != nil {
		return nil, err
	}

	log.Warnf("new value '%v'", valueName)

	return value, nil
}

func parseTagOrValueName(name string) string {
	buffer := new(bytes.Buffer)
	var escaped bool

	for _, r := range name {
		if escaped {
			buffer.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		default:
			buffer.WriteRune(r)
		}
	}

	return buffer.String()
}

func parseTagEqValueName(tagArg string) (string, string) {
	tagNameBuffer := new(bytes.Buffer)
	valueNameBuffer := new(bytes.Buffer)
	var buffer = tagNameBuffer
	var escaped bool

	for _, r := range tagArg {
		if escaped {
			buffer.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			escaped = true
		case '=':
			if buffer == tagNameBuffer {
				buffer = valueNameBuffer
			} else {
				buffer.WriteRune(r)
			}
		default:
			buffer.WriteRune(r)
		}
	}

	return tagNameBuffer.String(), valueNameBuffer.String()
}

func formatTagValueName(tagName, valueName string, useColour, implicit, explicit bool) string {
	tagName = escape(tagName, '=', ' ')
	valueName = escape(valueName, '=', ' ')

	if useColour {
		colourCode := colourCodeFor(implicit, explicit)

		if valueName == "" {
			return colourCode + tagName + ansi.ResetCode
		}

		return colourCode + tagName + ansi.ResetCode + "=" + colourCode + valueName + ansi.ResetCode
	}

	if valueName == "" {
		return tagName
	}

	return tagName + "=" + valueName
}

func colourCodeFor(implicit, explicit bool) string {
	if implicit && explicit {
		return ansi.YellowCode
	}

	if implicit {
		return ansi.CyanCode
	}

	return ""
}

func escape(text string, chars ...rune) string {
	for _, char := range chars {
		text = strings.Replace(text, string(char), `\`+string(char), -1)
	}

	return text
}
