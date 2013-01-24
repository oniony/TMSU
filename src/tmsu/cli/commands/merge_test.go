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

package commands

import (
	"os"
	"testing"
	"time"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/storage"
)

func TestMergeSingleTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileA, err := store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileA1, err := store.AddFile("/tmp/a/1", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB1, err := store.AddFile("/tmp/b/1", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagA, err := store.AddTag("a")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileA.Id, tagA.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileA1.Id, tagA.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileB1.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	command := MergeCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"a", "b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tagA, err = store.TagByName("a")
	if err != nil {
		test.Fatal(err)
	}
	if tagA != nil {
		test.Fatal("Tag 'a' still exists.")
	}

	tagB, err = store.TagByName("b")
	if err != nil {
		test.Fatal(err)
	}
	if tagB == nil {
		test.Fatal("Tag 'b' does not exist.")
	}

	expectExplicitTags(test, store, fileA, tagB)
	expectImplicitTags(test, store, fileA)
	expectExplicitTags(test, store, fileA1)
	expectImplicitTags(test, store, fileA1, tagB)
	expectExplicitTags(test, store, fileB, tagB)
	expectImplicitTags(test, store, fileB)
	expectExplicitTags(test, store, fileB1)
	expectImplicitTags(test, store, fileB1, tagB)
}

func TestMergeMultipleTags(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileA, err := store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileA1, err := store.AddFile("/tmp/a/1", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB1, err := store.AddFile("/tmp/b/1", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileC, err := store.AddFile("/tmp/c", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileC1, err := store.AddFile("/tmp/c/1", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagA, err := store.AddTag("a")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	tagC, err := store.AddTag("c")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileA.Id, tagA.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileA1.Id, tagA.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileB1.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileC.Id, tagC.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileC1.Id, tagC.Id); err != nil {
		test.Fatal(err)
	}

	command := MergeCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"a", "b", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	tagA, err = store.TagByName("a")
	if err != nil {
		test.Fatal(err)
	}
	if tagA != nil {
		test.Fatal("Tag 'a' still exists.")
	}

	tagB, err = store.TagByName("b")
	if err != nil {
		test.Fatal(err)
	}
	if tagB != nil {
		test.Fatal("Tag 'b' still exists.")
	}

	tagC, err = store.TagByName("c")
	if err != nil {
		test.Fatal(err)
	}
	if tagC == nil {
		test.Fatal("Tag 'c' does not exist.")
	}

	expectExplicitTags(test, store, fileA, tagC)
	expectImplicitTags(test, store, fileA)
	expectExplicitTags(test, store, fileA1)
	expectImplicitTags(test, store, fileA1, tagC)
	expectExplicitTags(test, store, fileB, tagC)
	expectImplicitTags(test, store, fileB)
	expectExplicitTags(test, store, fileB1)
	expectImplicitTags(test, store, fileB1, tagC)
	expectExplicitTags(test, store, fileC, tagC)
	expectImplicitTags(test, store, fileC)
	expectExplicitTags(test, store, fileC1)
	expectImplicitTags(test, store, fileC1, tagC)
}

func TestMergeNonExistentSourceTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	command := MergeCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"a", "b"}); err == nil {
		test.Fatal("Expected non-existent source tag to be identified.")
	}
}

func TestMergeNonExistentDestinationTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tagB, err := store.AddTag("a")
	if err != nil {
		test.Fatal(err)
	}

	command := MergeCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"a", "b"}); err == nil {
		test.Fatal("Expected non-existent destination tag to be identified.")
	}
}

func TestMergeSourceAndDestinationTheSame(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	tagB, err := store.AddTag("a")
	if err != nil {
		test.Fatal(err)
	}

	command := MergeCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"a", "a"}); err == nil {
		test.Fatal("Expected source and destination the same tag to be identified.")
	}
}
