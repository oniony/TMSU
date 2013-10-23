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
	"tmsu/log"
)

func TestStatusReport(test *testing.T) {
	// set-up

	databasePath := testDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := configureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	if err := createFile("/tmp/tmsu/a", "a"); err != nil {
		test.Fatalf("Could not create file: %v", err)
	}
	defer os.Remove("/tmp/tmsu/a")

	if err := createFile("/tmp/tmsu/b", "b"); err != nil {
		test.Fatalf("Could not create file: %v", err)
	}
	defer os.Remove("/tmp/tmsu/b")

	if err := createFile("/tmp/tmsu/c", "b"); err != nil {
		test.Fatalf("Could not create file: %v", err)
	}
	defer os.Remove("/tmp/tmsu/c")

	if err := createFile("/tmp/tmsu/d", "d"); err != nil {
		test.Fatalf("Could not create file: %v", err)
	}
	defer os.Remove("/tmp/tmsu/d")

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "a"}); err != nil {
		test.Fatal(err)
	}

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/b", "b"}); err != nil {
		test.Fatal(err)
	}

	if err := TagCommand.Exec(Options{}, []string{"/tmp/tmsu/d", "d"}); err != nil {
		test.Fatal(err)
	}

	if err := os.Remove("/tmp/tmsu/b"); err != nil {
		test.Fatal(err)
	}
	if err := createFile("/tmp/tmsu/b", "b"); err != nil {
		test.Fatalf("Could not create file: %v", err)
	}

	if err := os.Remove("/tmp/tmsu/d"); err != nil {
		test.Fatal(err)
	}

	// test

	if err := StatusCommand.Exec(Options{}, []string{"/tmp/tmsu/a", "/tmp/tmsu/b", "/tmp/tmsu/c", "/tmp/tmsu/d"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	compareOutput(test, "T /tmp/tmsu/a\nM /tmp/tmsu/b\n! /tmp/tmsu/d\nU /tmp/tmsu/c\n", string(bytes))
}
