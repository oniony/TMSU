package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tmsu/log"
)

func ConfigureOutput() (string, string, error) {
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

func ConfigureDatabase() string {
	databasePath := filepath.Join(os.TempDir(), "tmsu_test.db")
	os.Setenv("TMSU_DB", databasePath)

	return databasePath
}

func CompareOutput(test *testing.T, expected, actual string) {
	if actual != expected {
		test.Fatal("Output was not as expected.\nExpected: " + strings.Replace(expected, "\n", "\\n", -1) + "\nActual: " + strings.Replace(actual, "\n", "\\n", -1))
	}
}
