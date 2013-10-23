package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tmsu/entities"
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

func testDatabase() string {
	databasePath := filepath.Join(os.TempDir(), "tmsu_test.db")
	database.Path = databasePath
	return databasePath
}

func compareOutput(test *testing.T, expected, actual string) {
	if actual != expected {
		test.Fatal("Output was not as expected.\nExpected: " + strings.Replace(expected, "\n", "\\n", -1) + "\nActual: " + strings.Replace(actual, "\n", "\\n", -1))
	}
}

func expectTags(test *testing.T, store *storage.Storage, file *entities.File, tags ...*entities.Tag) {
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

func createFile(path string, contents string) error {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(contents)

	return nil
}
