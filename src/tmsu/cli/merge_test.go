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

func TestMergeSingleTag(test *testing.T) {
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

	fileA1, err := store.AddFile(tx, "/tmp/a/1", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	fileB1, err := store.AddFile(tx, "/tmp/b/1", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagA, err := store.AddTag(tx, "a")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag(tx, "b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagA.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA1.Id, tagA.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileB1.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := MergeCommand.Exec(store, Options{}, []string{"a", "b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	tagA, err = store.TagByName(tx, "a")
	if err != nil {
		test.Fatal(err)
	}
	if tagA != nil {
		test.Fatal("Tag 'a' still exists.")
	}

	tagB, err = store.TagByName(tx, "b")
	if err != nil {
		test.Fatal(err)
	}
	if tagB == nil {
		test.Fatal("Tag 'b' does not exist.")
	}

	expectTags(test, store, tx, fileA, tagB)
	expectTags(test, store, tx, fileA1, tagB)
	expectTags(test, store, tx, fileB, tagB)
	expectTags(test, store, tx, fileB1, tagB)
}

func TestMergeMultipleTags(test *testing.T) {
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

	fileA1, err := store.AddFile(tx, "/tmp/a/1", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	fileB1, err := store.AddFile(tx, "/tmp/b/1", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	fileC, err := store.AddFile(tx, "/tmp/c", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	fileC1, err := store.AddFile(tx, "/tmp/c/1", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagA, err := store.AddTag(tx, "a")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag(tx, "b")
	if err != nil {
		test.Fatal(err)
	}

	tagC, err := store.AddTag(tx, "c")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagA.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA1.Id, tagA.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileB1.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileC.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileC1.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := MergeCommand.Exec(store, Options{}, []string{"a", "b", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	tagA, err = store.TagByName(tx, "a")
	if err != nil {
		test.Fatal(err)
	}
	if tagA != nil {
		test.Fatal("Tag 'a' still exists.")
	}

	tagB, err = store.TagByName(tx, "b")
	if err != nil {
		test.Fatal(err)
	}
	if tagB != nil {
		test.Fatal("Tag 'b' still exists.")
	}

	tagC, err = store.TagByName(tx, "c")
	if err != nil {
		test.Fatal(err)
	}
	if tagC == nil {
		test.Fatal("Tag 'c' does not exist.")
	}

	expectTags(test, store, tx, fileA, tagC)
	expectTags(test, store, tx, fileA1, tagC)
	expectTags(test, store, tx, fileB, tagC)
	expectTags(test, store, tx, fileB1, tagC)
	expectTags(test, store, tx, fileC, tagC)
	expectTags(test, store, tx, fileC1, tagC)
}

func TestMergeNonExistentSourceTag(test *testing.T) {
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

	_, err = store.AddTag(tx, "b")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := MergeCommand.Exec(store, Options{}, []string{"a", "b"}); err == nil {
		test.Fatal("Expected non-existent source tag to be identified.")
	}
}

func TestMergeNonExistentDestinationTag(test *testing.T) {
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

	_, err = store.AddTag(tx, "a")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := MergeCommand.Exec(store, Options{}, []string{"a", "b"}); err == nil {
		test.Fatal("Expected non-existent destination tag to be identified.")
	}
}

func TestMergeSourceAndDestinationTheSame(test *testing.T) {
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

	_, err = store.AddTag(tx, "a")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := MergeCommand.Exec(store, Options{}, []string{"a", "a"}); err == nil {
		test.Fatal("Expected source and destination the same tag to be identified.")
	}
}
