package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

func configureOutput() (string, string, error) {
	outPath := filepath.Join(os.TempDir(), "tmsu_test.out")
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", "", fmt.Errorf("could not create output file '%v': %v", outPath, err)
	}
	log.Outfile = outFile

	errPath := filepath.Join(os.TempDir(), "tmsu_test.err")
	errFile, err := os.Create(errPath)
	if err != nil {
		return "", "", fmt.Errorf("could not create error file '%v': %v", outPath, err)
	}
	log.Errfile = errFile

	return outPath, errPath, nil
}

func configureDatabase() string {
	databasePath := filepath.Join(os.TempDir(), "tmsu_test.db")
	os.Setenv("TMSU_DB", databasePath)

	return databasePath
}

func compareOutput(test *testing.T, expected, actual string) {
	if actual != expected {
		test.Fatal("Output was not as expected.\nExpected: " + strings.Replace(expected, "\n", "\\n", -1) + "\nActual: " + strings.Replace(actual, "\n", "\\n", -1))
	}
}

func expectTags(test *testing.T, store *storage.Storage, file *database.File, tags ...*database.Tag) {
	fileTags, err := store.FileTagsByFileId(file.Id)
	if err != nil {
		test.Fatal(err)
	}
	if len(fileTags) != len(tags) {
		test.Fatalf("File '%v' has %v tags but expected %v.", file.Path(), len(fileTags), len(tags))
	}
	for index, filetag := range fileTags {
		tag := tags[index]

		if filetag.TagId != tag.Id {
			test.Fatal("File '%v' is tagged %v but expected %v.", file.Path(), filetag.TagId, tag.Id)
		}
	}
}
