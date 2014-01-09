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

package fingerprint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegularGeneration(test *testing.T) {
	tempFilePath := filepath.Join(os.TempDir(), "tmsu-fingerprint")

	file, err := os.Create(tempFilePath)
	if err != nil {
		test.Fatal(err.Error())
	}
	defer os.Remove(tempFilePath)

	_, err = file.WriteString("They were the footprints of a giagantic hound.")
	if err != nil {
		test.Fatal(err.Error())
	}

	fingerprint, err := Create(tempFilePath)
	if err != nil {
		test.Fatal(err.Error())
	}

	if fingerprint != Fingerprint("87d74123749a45e4c4e5e9053986d7ae878268a8e301d1b8125791517c0d39bf") {
		test.Fatal("Fingerprint incorrect.")
	}
}
