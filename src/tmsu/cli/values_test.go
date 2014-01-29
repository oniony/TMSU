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

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	file, err := store.AddFile("/tmp/tmsu/a", fingerprint.Fingerprint("123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	materialTag, err := store.AddTag("material")
	if err != nil {
		test.Fatal(err)
	}

	woodValue, err := store.AddValue("wood")
	if err != nil {
		test.Fatal(err)
	}

	metalValue, err := store.AddValue("metal")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(file.Id, materialTag.Id, woodValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(file.Id, materialTag.Id, metalValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	store.Commit()

	// test

	if err := ValuesCommand.Exec(Options{}, []string{"material"}); err != nil {
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

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	file, err := store.AddFile("/tmp/tmsu/a", fingerprint.Fingerprint("123"), time.Now(), 0, false)
	if err != nil {
		test.Fatal(err)
	}

	materialTag, err := store.AddTag("material")
	if err != nil {
		test.Fatal(err)
	}

	shapeTag, err := store.AddTag("shape")
	if err != nil {
		test.Fatal(err)
	}

	woodValue, err := store.AddValue("wood")
	if err != nil {
		test.Fatal(err)
	}

	metalValue, err := store.AddValue("metal")
	if err != nil {
		test.Fatal(err)
	}

	torroidValue, err := store.AddValue("torroid")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(file.Id, materialTag.Id, woodValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(file.Id, materialTag.Id, metalValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddFileTag(file.Id, shapeTag.Id, torroidValue.Id)
	if err != nil {
		test.Fatal(err)
	}

	store.Commit()

	// test

	if err := ValuesCommand.Exec(Options{}, []string{"material", "shape"}); err != nil {
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

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	woodValue, err := store.AddValue("wood")
	if err != nil {
		test.Fatal(err)
	}

	metalValue, err := store.AddValue("metal")
	if err != nil {
		test.Fatal(err)
	}

	torroidValue, err := store.AddValue("torroid")
	if err != nil {
		test.Fatal(err)
	}

	store.Commit()

	// test

	if err := ValuesCommand.Exec(Options{Option{"--all", "-a", "", false, ""}}, []string{}); err != nil {
		test.Fatal(err)
	}

	// verify

	outFile.Seek(0, 0)

	bytes, err := ioutil.ReadAll(outFile)
	compareOutput(test, "material: metal wood\nshape: torroid\n", string(bytes))
}
