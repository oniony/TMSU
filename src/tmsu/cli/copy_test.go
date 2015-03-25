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

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	fileAB, err := store.AddFile(tx, "/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	sourceTag, err := store.AddTag(tx, "source")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, sourceTag.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileAB.Id, sourceTag.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := CopyCommand.Exec(store, Options{}, []string{"source", "dest"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	destTag, err := store.TagByName(tx, "dest")
	if err != nil {
		test.Fatal(err)
	}
	if destTag == nil {
		test.Fatal("Destination tag does not exist.")
	}

	expectTags(test, store, tx, fileA, sourceTag, destTag)
	expectTags(test, store, tx, fileAB, sourceTag, destTag)
}

func TestCopyNonExistentSourceTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = CopyCommand.Exec(store, Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Non-existent source tag was not identified.")
	}
}

func TestCopyInvalidDestTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = CopyCommand.Exec(store, Options{}, []string{"source", "slash/invalid"})

	// validate

	if err == nil {
		test.Fatal("Invalid dest tag not identified.")
	}
}

func TestCopyDestTagAlreadyExists(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag(tx, "source")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag(tx, "dest")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	err = CopyCommand.Exec(store, Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Existing dest tag not identified.")
	}
}
