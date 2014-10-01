/*
Copyright 2011-2014 Paul Ruane.

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
	"tmsu/common/fingerprint"
	"tmsu/storage"
)

func TestCopySuccessful(test *testing.T) {
	// set-up

	databasePath := testDatabase()
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

	fileAB, err := store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	sourceTag, err := store.AddTag("source")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(fileA.Id, sourceTag.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(fileAB.Id, sourceTag.Id, 0); err != nil {
		test.Fatal(err)
	}

	// test

	if err := CopyCommand.Exec(Options{}, []string{"source", "dest"}); err != nil {
		test.Fatal(err)
	}

	// validate

	destTag, err := store.TagByName("dest")
	if err != nil {
		test.Fatal(err)
	}
	if destTag == nil {
		test.Fatal("Destination tag does not exist.")
	}

	expectTags(test, store, fileA, sourceTag, destTag)
	expectTags(test, store, fileAB, sourceTag, destTag)
}

func TestCopyNonExistentSourceTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	// test

	err := CopyCommand.Exec(Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Non-existent source tag was not identified.")
	}
}

func TestCopyInvalidDestTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	// test

	err := CopyCommand.Exec(Options{}, []string{"source", "slash/invalid"})

	// validate

	if err == nil {
		test.Fatal("Invalid dest tag not identified.")
	}
}

func TestCopyDestTagAlreadyExists(test *testing.T) {
	// set-up

	databasePath := testDatabase()
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

	// test

	err = CopyCommand.Exec(Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Existing dest tag not identified.")
	}
}
