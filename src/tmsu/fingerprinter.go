package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func Fingerprint(path string) (string, error) {
	file, error := os.Open(path)
	if error != nil {
		return "", error
	}
	defer file.Close()

	hash := sha256.New()

	buffer := make([]byte, 1024)
	for count := 0; error == nil; count, error = file.Read(buffer) {
		hash.Write(buffer[:count])
	}

	sum := hash.Sum()
	fingerprint := hex.EncodeToString(sum)

	return fingerprint, nil
}
