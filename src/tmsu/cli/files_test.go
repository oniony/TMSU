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

func TestFilesAll(test *testing.T) {
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

	_, err = store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	_, err = store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n/tmp/b/a\n/tmp/d\n", string(bytes))
}

func TestFilesSingleTag(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
	if err != nil {
		test.Fatal(err)
	}
	tagB, err := store.AddTag(tx, "b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n/tmp/b/a\n", string(bytes))
}

func TestFilesNotSingleTag(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}
	tagD, err := store.AddTag(tx, "d")
	if err != nil {
		test.Fatal(err)
	}
	tagB, err := store.AddTag(tx, "b")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"not", "b"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/d\n", string(bytes))
}

func TestFilesImplicitAnd(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
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

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b/a\n", string(bytes))
}

func TestFilesAnd(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
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

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b", "and", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b/a\n", string(bytes))
}

func TestFilesImplicitAndNot(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
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

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b", "not", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n", string(bytes))
}

func TestFilesAndNot(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
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

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b", "and", "not", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n", string(bytes))
}

func TestFilesOr(test *testing.T) {
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

	fileD, err := store.AddFile(tx, "/tmp/d", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileBA, err := store.AddFile(tx, "/tmp/b/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, true)
	if err != nil {
		test.Fatal(err)
	}

	tagD, err := store.AddTag(tx, "d")
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

	if _, err := store.AddFileTag(tx, fileD.Id, tagD.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagB.Id, 0); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileBA.Id, tagC.Id, 0); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"b", "or", "c"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n/tmp/b/a\n", string(bytes))
}

func TestFilesTagEqualsValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"size", "=", "100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size = 100"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n/tmp/b\n", string(bytes))
}

func TestFilesTagNotEqualsValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"not", "size", "=", "100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"not size = 100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"not size eq 100"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/a\n/tmp/a\n/tmp/a\n", string(bytes))
}

func TestFilesTagLessThanValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"size", "<", "100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size < 100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size lt 100"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/a\n/tmp/a\n/tmp/a\n", string(bytes))
}

func TestFilesTagGreaterThanValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"size", ">", "99"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size > 99"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size gt 99"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/b\n/tmp/b\n/tmp/b\n", string(bytes))
}

func TestFilesTagLessThanOrEqualToValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"size", "<=", "100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size <= 100"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size le 100"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/a\n/tmp/b\n/tmp/a\n/tmp/b\n/tmp/a\n/tmp/b\n", string(bytes))
}

func TestFilesTagGreaterThanOrEqualToValue(test *testing.T) {
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

	fileA, err := store.AddFile(tx, "/tmp/a", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}
	fileB, err := store.AddFile(tx, "/tmp/b", fingerprint.Fingerprint("abc"), time.Now(), 123, false)
	if err != nil {
		test.Fatal(err)
	}

	tagSize, err := store.AddTag(tx, "size")
	if err != nil {
		test.Fatal(err)
	}

	value99, err := store.AddValue(tx, "99")
	if err != nil {
		test.Fatal(err)
	}
	value100, err := store.AddValue(tx, "100")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddFileTag(tx, fileA.Id, tagSize.Id, value99.Id); err != nil {
		test.Fatal(err)
	}
	if _, err := store.AddFileTag(tx, fileB.Id, tagSize.Id, value100.Id); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := FilesCommand.Exec(store, Options{}, []string{"size", ">=", "99"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size >= 99"}); err != nil {
		test.Fatal(err)
	}
	if err := FilesCommand.Exec(store, Options{}, []string{"size ge 99"}); err != nil {
		test.Fatal(err)
	}

	// validate

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "/tmp/a\n/tmp/b\n/tmp/a\n/tmp/b\n/tmp/a\n/tmp/b\n", string(bytes))
}

//TODO tests for 'file' and 'directory' options.
