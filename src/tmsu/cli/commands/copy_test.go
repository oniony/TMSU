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
	"path/filepath"
	"testing"
	"time"
	"tmsu/cli"
	"tmsu/fingerprint"
	"tmsu/storage"
)

func TestSuccessfulCopy(test *testing.T) {
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

	fileB, err := store.AddFile("/tmp/a/b", fingerprint.Fingerprint("abc"), time.Now(), 123)
	if err != nil {
		test.Fatal(err)
	}

	fromTag, err := store.AddTag("from")
	if err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddExplicitFileTag(fileA.Id, fromTag.Id); err != nil {
		test.Fatal(err)
	}

	if _, err := store.AddImplicitFileTag(fileB.Id, fromTag.Id); err != nil {
		test.Fatal(err)
	}

	command := CopyCommand{false}

	// test

	if err := command.Exec(cli.Options{}, []string{"from", "to"}); err != nil {
		test.Fatal(err)
	}

	// validate

	toTag, err := store.TagByName("to")
	if err != nil {
		test.Fatal(err)
	}
	if toTag == nil {
		test.Fatal("Destination tag does not exist.")
	}

	explicitFileTagsA, err := store.ExplicitFileTagsByFileId(fileA.Id)
	if err != nil {
		test.Fatal(err)
	}
	if len(explicitFileTagsA) != 2 {
		test.Fatal("Expected two file-tags for file 'a'.")
	}
	if explicitFileTagsA[0].TagId != fromTag.Id {
		test.Fatalf("Explicit file-tag for from tag has wrong tag ID '%v'.", explicitFileTagsA[0].TagId)
	}
	if explicitFileTagsA[1].TagId != toTag.Id {
		test.Fatalf("Explicit file-tag for to tag has wrong tag ID '%v'.", explicitFileTagsA[1].TagId)
	}

	implicitFileTagsA, err := store.ImplicitFileTagsByFileId(fileB.Id)
	if err != nil {
		test.Fatal(err)
	}
	if len(implicitFileTagsA) != 2 {
		test.Fatal("impected two file-tags for file 'a'.")
	}
	if implicitFileTagsA[0].TagId != fromTag.Id {
		test.Fatalf("implicit file-tag for from tag has wrong tag ID '%v'.", implicitFileTagsA[0].TagId)
	}
	if implicitFileTagsA[1].TagId != toTag.Id {
		test.Fatalf("implicit file-tag for to tag has wrong tag ID '%v'.", implicitFileTagsA[1].TagId)
	}
}

func TestNonExistentFromTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	command := CopyCommand{false}

	// test

	err := command.Exec(cli.Options{}, []string{"from", "to"})

	// validate

	if err == nil {
		test.Fatal("Non-existent from tag was not identified.")
	}
}

func TestInvalidToTag(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	command := CopyCommand{false}

	// test

	err := command.Exec(cli.Options{}, []string{"from", "to/dest"})

	// validate

	if err == nil {
		test.Fatal("Invalid to tag not identified.")
	}
}

func TestToTagAlreadyExists(test *testing.T) {
	// set-up

	databasePath := configureDatabase()
	defer os.Remove(databasePath)

	store, err := storage.Open()
	if err != nil {
		test.Fatal(err)
	}
	defer store.Close()

	_, err = store.AddTag("from")
	if err != nil {
		test.Fatal(err)
	}

	_, err = store.AddTag("to")
	if err != nil {
		test.Fatal(err)
	}

	command := CopyCommand{false}

	// test

	err = command.Exec(cli.Options{}, []string{"from", "to"})

	// validate

	if err == nil {
		test.Fatal("Existing to tag not identified.")
	}
}

//-

func configureDatabase() string {
	databasePath := filepath.Join(os.TempDir(), "tmsu_test.db")
	os.Setenv("TMSU_DB", databasePath)

	return databasePath
}
