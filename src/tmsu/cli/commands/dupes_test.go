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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
)

func TestSingleDupe(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "Set of 2 duplicates:\n  /tmp/a\n  /tmp/a/b\n", string(bytes))
}

func TestMultipleDupes(test *testing.T) {
	// set-up
	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("def"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/e/f", fingerprint.Fingerprint("def"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/d", fingerprint.Fingerprint("def"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "Set of 2 duplicates:\n  /tmp/a\n  /tmp/a/b\n\nSet of 3 duplicates:\n  /tmp/a/d\n  /tmp/b\n  /tmp/e/f\n", string(bytes))
}

func TestNoDupes(test *testing.T) {
	// set-up
	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("def"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("ghi"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/e/f", fingerprint.Fingerprint("jkl"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/d", fingerprint.Fingerprint("mno"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "", string(bytes))
}

func TestSingleDupeOfFile(test *testing.T) {
	// set-up
	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()

	path := filepath.Join(os.TempDir(), "tmsu-file")
	_, err = os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(path)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("xxx"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/e/f", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/d", fingerprint.Fingerprint("xxx"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{path}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/a\n", string(bytes))
}

func TestMultipleDupeOfFile(test *testing.T) {
	// set-up
	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()

	path := filepath.Join(os.TempDir(), "tmsu-file")
	_, err = os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(path)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("xxx"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/e/f", fingerprint.Fingerprint("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/d", fingerprint.Fingerprint("xxx"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{path}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/a\n/tmp/a/b\n/tmp/e/f\n", string(bytes))
}

func TestNoDupeOfFile(test *testing.T) {
	// set-up
	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	err := ConfigureOutput()

	path := filepath.Join(os.TempDir(), "tmsu-file")
	_, err = os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(path)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/b", fingerprint.Fingerprint("def"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("ghi"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/e/f", fingerprint.Fingerprint("klm"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile("/tmp/a/d", fingerprint.Fingerprint("nop"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := DupesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{path}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "", string(bytes))
}
