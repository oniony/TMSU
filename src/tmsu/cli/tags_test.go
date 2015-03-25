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
	"io/ioutil"
	"os"
	"testing"
	"time"
	"tmsu/common/fingerprint"
	"tmsu/storage"
)

func TestTagsForSingleFile(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	err := redirectStreams()
	if err != nil {
		test.Fatal(err)
	}
	defer restoreStreams()

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	file, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("123"), time.Now(), 0, false)
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

	if err := TagsCommand.Exec(store, Options{}, []string{"/tmp/tmsu/a"}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/tmsu/a: apple banana\n", string(bytes))
}

func TestTagsForMultipleFiles(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	err := redirectStreams()
	if err != nil {
		test.Fatal(err)
	}
	defer restoreStreams()

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	aFile, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	bFile, err := store.AddFile(tx, "/tmp/tmsu/b", fingerprint.Fingerprint("123"), time.Now(), 0, false)
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

	_, err = store.AddFileTag(tx, aFile.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, aFile.Id, bananaTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, bFile.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := TagsCommand.Exec(store, Options{}, []string{"/tmp/tmsu/a", "/tmp/tmsu/b"}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/tmsu/a: apple banana\n/tmp/tmsu/b: apple\n", string(bytes))
}

func TestAllTags(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	err := redirectStreams()
	if err != nil {
		test.Fatal(err)
	}
	defer restoreStreams()

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag(tx, "banana")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := TagsCommand.Exec(store, Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "apple\nbanana\n", string(bytes))
}

func TestImpliedTags(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	err := redirectStreams()
	if err != nil {
		test.Fatal(err)
	}
	defer restoreStreams()

	store, err := storage.OpenAt(databasePath)
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		test.Fatal(err)
	}

	file, err := store.AddFile(tx, "/tmp/tmsu/a", fingerprint.Fingerprint("123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	appleTag, err := store.AddTag(tx, "apple")
	if err != nil {
		test.Fatal(err)
	}

	fruitTag, err := store.AddTag(tx, "fruit")
	if err != nil {
		test.Fatal(err)
	}

	foodTag, err := store.AddTag(tx, "food")
	if err != nil {
		test.Fatal(err)
	}

	if err := store.AddImplication(tx, appleTag.Id, fruitTag.Id); err != nil {
		test.Fatal(err)
	}

	if err := store.AddImplication(tx, fruitTag.Id, foodTag.Id); err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, appleTag.Id, 0)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := TagsCommand.Exec(store, Options{}, []string{"/tmp/tmsu/a"}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/tmsu/a: apple food fruit\n", string(bytes))
}
