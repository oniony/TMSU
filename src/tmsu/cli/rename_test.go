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

func TestRenameTag(test *testing.T) {
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

	fileAB, err := store.AddFile(tx, "/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
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

	if err := RenameCommand.Exec(store, Options{}, []string{"source", "dest"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	originalTag, err := store.TagByName(tx, "source")
	if err != nil {
		test.Fatal(err)
	}
	if originalTag != nil {
		test.Fatal("Tag with original name still exists.")
	}

	destTag, err := store.TagByName(tx, "dest")
	if err != nil {
		test.Fatal(err)
	}
	if destTag == nil {
		test.Fatal("Destination tag does not exist.")
	}
	if destTag.Id != sourceTag.Id {
		test.Fatalf("Renamed tag has different ID.")
	}

	expectTags(test, store, tx, fileA, sourceTag)
	expectTags(test, store, tx, fileAB, sourceTag)
}

func TestRenameNonExistentSourceTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = RenameCommand.Exec(store, Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Non-existent source tag was not identified.")
	}
}

func TestRenameInvalidDestTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = RenameCommand.Exec(store, Options{}, []string{"source", "slash/invalid"})

	// validate

	if err == nil {
		test.Fatal("Invalid dest tag not identified.")
	}
}

func TestRenameExistingDestTag(test *testing.T) {
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

	err = RenameCommand.Exec(store, Options{}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Existing dest tag not identified.")
	}
}

func TestRenameValue(test *testing.T) {
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

	fileAB, err := store.AddFile(tx, "/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tag, err := store.AddTag(tx, "tag")
	if err != nil {
		test.Fatal(err)
	}

	sourceValue, err := store.AddValue(tx, "source")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tag.Id, sourceValue.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileAB.Id, tag.Id, sourceValue.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := RenameCommand.Exec(store, Options{{"--value", "", "", false, ""}}, []string{"source", "dest"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	originalValue, err := store.ValueByName(tx, "source")
	if err != nil {
		test.Fatal(err)
	}
	if originalValue != nil {
		test.Fatal("Value with original name still exists.")
	}

	destValue, err := store.ValueByName(tx, "dest")
	if err != nil {
		test.Fatal(err)
	}
	if destValue == nil {
		test.Fatal("Destination value does not exist.")
	}
	if destValue.Id != sourceValue.Id {
		test.Fatalf("Renamed value has different ID.")
	}

	expectValue(test, store, tx, fileA, tag, sourceValue)
	expectValue(test, store, tx, fileAB, tag, sourceValue)
}

func TestRenameNonExistentSourceValue(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = RenameCommand.Exec(store, Options{{"--value", "", "", false, ""}}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Non-existent source value was not identified.")
	}
}

func TestRenameInvalidDestValue(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = RenameCommand.Exec(store, Options{{"--value", "", "", false, ""}}, []string{"source", "slash/invalid"})

	// validate

	if err == nil {
		test.Fatal("Invalid dest value not identified.")
	}
}

func TestRenameExistingDestValue(test *testing.T) {
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

	_, err = store.AddValue(tx, "source")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddValue(tx, "dest")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	err = RenameCommand.Exec(store, Options{{"--value", "", "", false, ""}}, []string{"source", "dest"})

	// validate

	if err == nil {
		test.Fatal("Existing dest value not identified.")
	}
}
