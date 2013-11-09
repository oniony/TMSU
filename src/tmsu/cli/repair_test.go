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
	"io/ioutil"
	"os"
	"testing"
	"tmsu/storage"
)

func TestRepairMovedFile(test *testing.T) {
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

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "a"}); err != nil {
		test.Fatal(err)
	}

	if err := os.Rename("/tmp/tmsu/a", "/tmp/tmsu/b"); err != nil {
		test.Fatal(err)
	}

	// test

	if err := RepairCommand.Exec(Options{}, []string{"/tmp/tmsu"}); err != nil {
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

	if files[0].Path() != "/tmp/tmsu/b" {
		test.Fatalf("File rename was not repaired.")
	}
}

func TestRepairModifiedFile(test *testing.T) {
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

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "a"}); err != nil {
		test.Fatal(err)
	}

	if err := createFile("/tmp/tmsu/a", "banana"); err != nil {
		test.Fatal(err)
	}

	// test

	if err := RepairCommand.Exec(Options{}, []string{"/tmp/tmsu"}); err != nil {
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

	if files[0].Size != 6 {
		test.Fatalf("File modification was not repaired.")
	}
}

func TestReportsMissingFiles(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	err := redirectStreams()
	if err != nil {
		test.Fatal(err)
	}
	defer restoreStreams()

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	if err := createFile("/tmp/tmsu/a", "hello"); err != nil {
		test.Fatal(err)
	}

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "a"}); err != nil {
		test.Fatal(err)
	}

	if err := os.Remove("/tmp/tmsu/a"); err != nil {
		test.Fatal(err)
	}

	// test

	if err := RepairCommand.Exec(Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "tmsu: /tmp/tmsu/a: missing\n", string(bytes))
}
