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

func TestValuesForSingleTag(test *testing.T) {
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

	materialTag, err := store.AddTag(tx, "material")
	if err != nil {
		test.Fatal(err)
	}

	woodValue, err := store.AddValue(tx, "wood")
	if err != nil {
		test.Fatal(err)
	}

	metalValue, err := store.AddValue(tx, "metal")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, materialTag.Id, woodValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, materialTag.Id, metalValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := ValuesCommand.Exec(store, Options{}, []string{"material"}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "metal\nwood\n", string(bytes))
}

func TestValuesForMulitpleTags(test *testing.T) {
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

	materialTag, err := store.AddTag(tx, "material")
	if err != nil {
		test.Fatal(err)
	}

	shapeTag, err := store.AddTag(tx, "shape")
	if err != nil {
		test.Fatal(err)
	}

	woodValue, err := store.AddValue(tx, "wood")
	if err != nil {
		test.Fatal(err)
	}

	metalValue, err := store.AddValue(tx, "metal")
	if err != nil {
		test.Fatal(err)
	}

	torroidValue, err := store.AddValue(tx, "torroid")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, materialTag.Id, woodValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, materialTag.Id, metalValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(tx, file.Id, shapeTag.Id, torroidValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := ValuesCommand.Exec(store, Options{}, []string{"material", "shape"}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "material: metal wood\nshape: torroid\n", string(bytes))
}

func TestAllValues(test *testing.T) {
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

	_, err = store.AddValue(tx, "wood")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddValue(tx, "metal")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddValue(tx, "torroid")
	if err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	// test

	if err := ValuesCommand.Exec(store, Options{}, []string{}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "metal\ntorroid\nwood\n", string(bytes))
}
