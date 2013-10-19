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

package cli

import (
	"os"
	"testing"
	"time"
	"tmsu/fingerprint"
	"tmsu/storage"
)

func TestRenameSuccessful(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileA, err := store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	fileAB, err := store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	sourceTag, err := store.AddTag("source")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(fileA.Id, sourceTag.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(fileAB.Id, sourceTag.Id); err != nil {
		test.Fatal(err)
	}

	command := renameCommand{false}

	// test

	if err := command.Exec(Options{}, []string{"source", "dest"}); err != nil {
		test.Fatal(err)
	}

	// validate

	originalTag, err := store.TagByName("source")
	if err != nil {
		test.Fatal(err)
	}
	if originalTag != nil {
		test.Fatal("Tag with original name still exists.")
	}

	destTag, err := store.TagByName("dest")
	if err != nil {
		test.Fatal(err)
	}
	if destTag == nil {
		test.Fatal("Destination tag does not exist.")
	}
	if destTag.Id != sourceTag.Id {
		test.Fatalf("Renamed tag has different ID.")
	}

	expectTags(test, store, fileA, sourceTag)
	expectTags(test, store, fileAB, sourceTag)
}

func TestRenameNonExistentSourceTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	command := renameCommand{false}

	// test

	err := command.Exec(Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Non-existent source tag was not identified.")
	}
}

func TestRenameInvalidDestTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	command := renameCommand{false}

	// test

	err := command.Exec(Options{}, []string{"source", "slash/invalid"})

	// validate

	if err == nil {
		test.Fatal("Invalid dest tag not identified.")
	}
}

func TestRenameDestTagAlreadyExists(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddTag("source")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag("dest")
	if err != nil {
		test.Fatal(err)
	}

	command := renameCommand{false}

	// test

	err = command.Exec(Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Existing dest tag not identified.")
	}
}
