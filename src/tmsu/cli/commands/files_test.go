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
	"testing"
	"time"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/log"
	"tmsu/storage"
)

func TestAllFiles(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{cli.Option{"-a", "--all", ""}}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b\n/tmp/b/a\n/tmp/d\n", string(bytes))
}

func TestSingleTag(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b\n/tmp/b/a\n", string(bytes))
}

func TestNotSingleTag(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"-b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/d\n", string(bytes))
}

func TestAnd(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
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

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileBA.Id, tagC.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"b", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b/a\n", string(bytes))
}

func TestAndNot(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
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

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileBA.Id, tagC.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"b", "-c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b\n", string(bytes))
}

func TestSingleTagExplicit(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{cli.Option{"--explicit", "-e", ""}}, []string{"b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b\n", string(bytes))
}

func TestNotSingleTagExplicit(test *testing.T) {
	// set-up

	databasePath := ConfigureDatabase()
	defer os.Remove(databasePath)

	outPath, errPath, err := ConfigureOutput()
	if err != nil {
		test.Fatal(err)
	}
	defer os.Remove(outPath)
	defer os.Remove(errPath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	fileD, err := store.AddFile("/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileBA, err := store.AddFile("/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fileB, err := store.AddFile("/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag("d")
	if err != nil {
		test.Fatal(err)
	}

	tagB, err := store.AddTag("b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileD.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileB.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileBA.Id, tagB.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileBA.Id, tagD.Id); err != nil {
		test.Fatal(err)
	}

	command := FilesCommand{false}

	// test

	if err := command.Exec(cli.Options{cli.Option{"--explicit", "-e", ""}}, []string{"-d"}); err != nil {
		test.Fatal(err)
	}

	// validate

	log.Outfile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(log.Outfile)
	CompareOutput(test, "/tmp/b\n", string(bytes))
}
