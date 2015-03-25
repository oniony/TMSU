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

func TestSingleUntag(test *testing.T) {
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

	file, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	appleTag, err := store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	bananaTag, err := store.AddTag(tx, "banana")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, bananaTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := UntagCommand.Exec(store, Options{}, []string{"/tmp/tmsu/a", "apple"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	fileTags, err := store.FileTags(tx)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 1 {
		test.Fatalf("Expected one file-tag but are %v", len(fileTags))
	}
	if fileTags[0].TagId != bananaTag.Id {
		test.Fatalf("Incorrect tag was applied.")
	}
}

func TestMultipleUntag(test *testing.T) {
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

	file, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	appleTag, err := store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	bananaTag, err := store.AddTag(tx, "banana")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, bananaTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := UntagCommand.Exec(store, Options{}, []string{"/tmp/tmsu/a", "apple", "banana"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	fileTags, err := store.FileTags(tx)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 0 {
		test.Fatalf("Expected no file-tags but are %v", len(fileTags))
	}

	files, err := store.Files(tx, "")
	if err != nil {
		test.Fatal(err)
	}
	if len(files) != 0 {
		test.Fatalf("Expected no files but are %v", len(files))
	}
}

func TestUntagMultipleFiles(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile(tx, "/tmp/tmsu/b", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	appleTag, err := store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, fileA.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, fileB.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := UntagCommand.Exec(store, Options{Option{"--tags", "-t", "", true, "apple"}}, []string{"/tmp/tmsu/a", "/tmp/tmsu/b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	fileTags, err := store.FileTags(tx)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 0 {
		test.Fatalf("Expected no file-tags but are %v", len(fileTags))
	}

	files, err := store.Files(tx, "")
	if err != nil {
		test.Fatal(err)
	}
	if len(files) != 0 {
		test.Fatalf("Expected no files but are %v", len(files))
	}
}

func TestUntagAll(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile(tx, "/tmp/tmsu/b", fingerprint.Fingerprint("abc123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	appleTag, err := store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, fileA.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, fileB.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := UntagCommand.Exec(store, Options{Option{"--all", "-a", "", false, ""}}, []string{"/tmp/tmsu/a", "/tmp/tmsu/b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	fileTags, err := store.FileTags(tx)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 0 {
		test.Fatalf("Expected no file-tags but are %v", len(fileTags))
	}

	files, err := store.Files(tx, "")
	if err != nil {
		test.Fatal(err)
	}
	if len(files) != 0 {
		test.Fatalf("Expected no files but are %v", len(files))
	}
}
