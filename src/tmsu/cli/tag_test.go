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
	"tmsu/storage"
)

func TestSingleTag(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	if err := createFile("/tmp/tmsu/a", "hello"); err != nil {
		test.Fatal(err)
	}
	defer os.Remove("/tmp/tmsu/a")

	store.Commit()

	// test

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "apple"}); err != nil {
		test.Fatal(err)
	}

	// validate

	files, err := store.Files()
	if err != nil {
		test.Fatal(err)
	}
	if len(files) != 1 {
		test.Fatalf("Expected one file but are %v", len(files))
	}
	if files[0].Path() != "/tmp/tmsu/a" {
		test.Fatalf("Incorrect file was added.")
	}

	tags, err := store.Tags()
	if err != nil {
		test.Fatal(err)
	}
	if len(tags) != 1 {
		test.Fatalf("Expected one tag but are %v", len(tags))
	}
	if tags[0].Name != "apple" {
		test.Fatalf("Incorrect tag was added.")
	}

	fileTags, err := store.FileTags()
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 1 {
		test.Fatalf("Expected one file-tag but are %v", len(fileTags))
	}
	if fileTags[0].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[0].TagId != tags[0].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
}

func TestMultipleTags(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	if err := createFile("/tmp/tmsu/a", "hello"); err != nil {
		test.Fatal(err)
	}
	defer os.Remove("/tmp/tmsu/a")

	store.Commit()

	// test

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "apple", "banana", "clementine"}); err != nil {
		test.Fatal(err)
	}

	// validate

	files, err := store.Files()
	if err != nil {
		test.Fatal(err)
	}
	if len(files) != 1 {
		test.Fatalf("Expected one file but are %v", len(files))
	}
	if files[0].Path() != "/tmp/tmsu/a" {
		test.Fatalf("Incorrect file was added.")
	}

	tags, err := store.Tags()
	if err != nil {
		test.Fatal(err)
	}
	if len(tags) != 3 {
		test.Fatalf("Expected three tags but are %v", len(tags))
	}
	if tags[0].Name != "apple" {
		test.Fatalf("Incorrect tag was added.")
	}
	if tags[1].Name != "banana" {
		test.Fatalf("Incorrect tag was added.")
	}
	if tags[2].Name != "clementine" {
		test.Fatalf("Incorrect tag was added.")
	}

	fileTags, err := store.FileTags()
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 3 {
		test.Fatalf("Expected three file-tags but are %v", len(fileTags))
	}
	if fileTags[0].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[0].TagId != tags[0].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
	if fileTags[1].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[1].TagId != tags[1].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
	if fileTags[2].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[2].TagId != tags[2].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
}

func TestTagMultipleFiles(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	if err := createFile("/tmp/tmsu/a", "hello"); err != nil {
		test.Fatal(err)
	}
	defer os.Remove("/tmp/tmsu/a")

	if err := createFile("/tmp/tmsu/b", "there"); err != nil {
		test.Fatal(err)
	}
	defer os.Remove("/tmp/tmsu/b")

	store.Commit()

	// test

	if err := TagCommand.Exec(Options{Option{"--tags", "-t", "", true, "apple banana"}}, []string{"/tmp/tmsu/a", "/tmp/tmsu/b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	files, err := store.Files()
	if err != nil {
		test.Fatal(err)
	}

	if len(files) != 2 {
		test.Fatalf("Expected two files but are %v", len(files))
	}
	if files[0].Path() != "/tmp/tmsu/a" {
		test.Fatalf("Incorrect file was added.")
	}
	if files[1].Path() != "/tmp/tmsu/b" {
		test.Fatalf("Incorrect file was added.")
	}

	tags, err := store.Tags()
	if err != nil {
		test.Fatal(err)
	}
	if len(tags) != 2 {
		test.Fatalf("Expected two tags but are %v", len(tags))
	}
	if tags[0].Name != "apple" {
		test.Fatalf("Incorrect tag was added.")
	}
	if tags[1].Name != "banana" {
		test.Fatalf("Incorrect tag was added.")
	}

	fileTags, err := store.FileTags()
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != 4 {
		test.Fatalf("Expected four file-tag but are %v", len(fileTags))
	}
	if fileTags[0].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[0].TagId != tags[0].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
	if fileTags[1].FileId != files[0].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[1].TagId != tags[1].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
	if fileTags[2].FileId != files[1].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[2].TagId != tags[0].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
	if fileTags[3].FileId != files[1].Id {
		test.Fatalf("Incorrect file was tagged.")
	}
	if fileTags[3].TagId != tags[1].Id {
		test.Fatalf("Incorrect tag was applied.")
	}
}

//TODO recursive
