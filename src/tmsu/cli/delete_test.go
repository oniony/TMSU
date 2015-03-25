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

func TestDeleteUnappliedTag(test *testing.T) {
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

	tagBeetroot, err := store.AddTag(tx, "beetroot")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := DeleteCommand.Exec(store, Options{}, []string{"beetroot"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	tagBeetroot, err = store.TagByName(tx, "beetroot")
	if err != nil {
		test.Fatal(err)
	}
	if tagBeetroot != nil {
		test.Fatal("Deleted tag still exists.")
	}
}

func TestDeleteAppliedTag(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	fileF, err := store.AddFile(tx, "/tmp/f", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagDeathrow, err := store.AddTag(tx, "deathrow")
	if err != nil {
		test.Fatal(err)
	}

	tagFreeman, err := store.AddTag(tx, "freeman")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileD.Id, tagDeathrow.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileF.Id, tagFreeman.Id, 0); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileB.Id, tagDeathrow.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagFreeman.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := DeleteCommand.Exec(store, Options{}, []string{"deathrow"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tx, err = store.Begin()
	if err != nil {
		test.Fatal(err)
	}
	defer tx.Commit()

	tagDeathrow, err = store.TagByName(tx, "deathrow")
	if err != nil {
		test.Fatal(err)
	}
	if tagDeathrow != nil {
		test.Fatal("Deleted tag still exists.")
	}

	fileTagsD, err := store.FileTagsByFileId(tx, fileD.Id, true)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTagsD) != 0 {
		test.Fatal("Expected no file-tags for file 'd'.")
	}

	fileTagsF, err := store.FileTagsByFileId(tx, fileF.Id, true)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTagsF) != 1 {
		test.Fatal("Expected one file-tag for file 'f'.")
	}
	if fileTagsF[0].TagId != tagFreeman.Id {
		test.Fatal("Expected file-tag for tag 'freeman'.")
	}

	fileTagsB, err := store.FileTagsByFileId(tx, fileB.Id, true)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTagsB) != 1 {
		test.Fatal("Expected one file-tag for file 'b'.")
	}
	if fileTagsB[0].TagId != tagFreeman.Id {
		test.Fatal("Expected file-tag for tag 'freeman'.")
	}
}

func TestDeleteNonExistentTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	// test

	err = DeleteCommand.Exec(store, Options{}, []string{"deleteme"})

	// validate

	if err == nil {
		test.Fatal("Non-existent from tag was not identified.")
	}
}
