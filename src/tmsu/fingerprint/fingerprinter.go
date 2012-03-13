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

package fingerprint

import (
    "bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"tmsu/common"
)

func Create(path string) (Fingerprint, error) {
	file, err := os.Open(path)
	if err != nil {
		return EMPTY, err
	}
	defer file.Close()

	hash := sha256.New()

    if (common.IsDir(path)) {
        dir, err := os.Open(path)
        if err != nil {
            return EMPTY, err
        }

        entries, err := dir.Readdir(0)
        if err != nil {
            return EMPTY, err
        }
        sort.Sort(FileInfoSlice(entries))

        buffer := new(bytes.Buffer)
        for _, entry := range entries {
            entryPath := filepath.Join(path, entry.Name())
            stat, err := os.Stat(entryPath)

            if err != nil {
                return EMPTY, err
            }

            binary.Write(buffer, binary.LittleEndian, stat.ModTime().UnixNano())
        }
        buffer.WriteTo(hash)
    } else {
        buffer := make([]byte, 1024)
        for count := 0; err == nil; count, err = file.Read(buffer) {
            hash.Write(buffer[:count])
        }
    }

    sum := hash.Sum(make([]byte, 0, 64))
    fingerprint := hex.EncodeToString(sum)

	return Fingerprint(fingerprint), nil
}

type FileInfoSlice []os.FileInfo

func (infos FileInfoSlice) Len() int {
    return len(infos)
}

func (infos FileInfoSlice) Less(i, j int) bool {
    return infos[i].Name() < infos[j].Name()
}

func (infos FileInfoSlice) Swap(i, j int) {
    infos[j], infos[i] = infos[i], infos[j]
}
