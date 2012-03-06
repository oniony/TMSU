/*
Copyright 2011-2012 Paul Ruane.

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

package common

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
)

func Fingerprint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

    if (IsDir(path)) {
        children, err := file.Readdir(0)
        if err != nil {
            return "", err
        }

        fingerprints := make([]string, 0, len(children))

        for _, child := range children {
            childPath := filepath.Join(path, child.Name())

            fingerprint, err := Fingerprint(childPath)
            if err != nil {
                return "", err
            }

            fingerprints = append(fingerprints, fingerprint)
        }

        sort.Strings(fingerprints)

        for _, fingerprint := range fingerprints {
            hash.Write([]byte(fingerprint))
        }
    } else {
        buffer := make([]byte, 1024)
        for count := 0; err == nil; count, err = file.Read(buffer) {
            hash.Write(buffer[:count])
        }

    }

    sum := hash.Sum(make([]byte, 0, 64))
    fingerprint := hex.EncodeToString(sum)

	return fingerprint, nil
}
