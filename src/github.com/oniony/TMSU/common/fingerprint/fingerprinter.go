// Copyright 2011-2017 Paul Ruane.

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

package fingerprint

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/blake2b"
)

const sparseFingerprintThreshold = 5 * 1024 * 1024
const sparseFingerprintSize = 512 * 1024

func Create(path, fileAlgorithm, directoryAlgorithm, symlinkAlgorithm string) (Fingerprint, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return Empty, err
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		if symlinkAlgorithm == "follow" {
			stat, err = os.Stat(path)
			if err != nil {
				return Empty, err
			}
		} else {
			return createSymlinkFingerprint(path, symlinkAlgorithm)
		}
	}

	switch {
	case stat.Mode().IsDir():
		return createDirectoryFingerprint(path, directoryAlgorithm)
	case stat.Mode().IsRegular():
		return createFileFingerprint(path, fileAlgorithm, stat)
	default:
		return Empty, fmt.Errorf("unsupported file mode '%v'", stat.Mode())
	}
}

// unexported

func createFileFingerprint(path, algorithm string, stat os.FileInfo) (Fingerprint, error) {
	switch algorithm {
	case "dynamic:SHA256", "":
		return dynamicFingerprint(path, sha256.New(), stat.Size())
	case "dynamic:SHA1":
		return dynamicFingerprint(path, sha1.New(), stat.Size())
	case "dynamic:MD5":
		return dynamicFingerprint(path, md5.New(), stat.Size())
	case "dynamic:BLAKE2b":
		hash, err := blake2b.New256(nil)
		if err != nil {
			// Should never happen actually.
			return "", err
		}
		return dynamicFingerprint(path, hash, stat.Size())
	case "SHA256":
		return regularFingerprint(path, sha256.New())
	case "SHA1":
		return regularFingerprint(path, sha1.New())
	case "MD5":
		return regularFingerprint(path, md5.New())
	case "BLAKE2b":
		hash, err := blake2b.New256(nil)
		if err != nil {
			// Should never happen actually.
			return "", err
		}
		return regularFingerprint(path, hash)
	case "none":
		return Empty, nil
	default:
		return "", fmt.Errorf("unsupported file fingerprint algorithm '%v'", algorithm)
	}
}

func createDirectoryFingerprint(path, algorithm string) (Fingerprint, error) {
	switch algorithm {
	case "sumSizes":
		return sumSizesFingerprint(path, 0)
	case "dynamic:sumSizes", "":
		return sumSizesFingerprint(path, 500)
	case "none":
		return Empty, nil
	default:
		return "", fmt.Errorf("unsupported directory fingerprint algorithm '%v'", algorithm)
	}
}

func createSymlinkFingerprint(path, algorithm string) (Fingerprint, error) {
	switch algorithm {
	case "targetName":
		return symlinkTargetNameFingerprint(path, true)
	case "targetNameNoExt":
		return symlinkTargetNameFingerprint(path, false)
	case "none":
		return Empty, nil
	default:
		return Empty, fmt.Errorf("unsupported symbolic link fingerprint algorithm '%v'", algorithm)
	}
}

func regularFingerprint(path string, h hash.Hash) (Fingerprint, error) {
	return calculateRegularFingerprint(path, h)
}

func dynamicFingerprint(path string, h hash.Hash, fileSize int64) (Fingerprint, error) {
	if fileSize > sparseFingerprintThreshold {
		return calculateSparseFingerprint(path, fileSize, h)
	}

	return calculateRegularFingerprint(path, h)
}

// Uses the symbolic target's filename as the fingerprint
func symlinkTargetNameFingerprint(path string, includeExtension bool) (Fingerprint, error) {
	target, err := os.Readlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Empty, nil
		}

		return Empty, fmt.Errorf("'%v': could not determine target of symbolic link: %v", path, err)
	}

	fingerprint := filepath.Base(target)

	if !includeExtension {
		pos := strings.Index(fingerprint, ".")
		if pos > -1 {
			fingerprint = fingerprint[0:pos]
		}
	}

	return Fingerprint(fingerprint), nil
}

// Creates a crude directory fingerprint by add the size of the contained files
func sumSizesFingerprint(path string, maxFiles uint) (Fingerprint, error) {
	paths := []string{path}
	var fileCount uint
	var totalSize int64

out:
	for index := 0; index < len(paths); index++ {
		path := paths[index]
		stats := stats(path)

		for _, stat := range stats {
			if stat.IsDir() {
				childPath := filepath.Join(path, stat.Name())
				paths = append(paths, childPath)
			} else {
				totalSize += stat.Size()

				if fileCount++; maxFiles != 0 && fileCount >= maxFiles {
					break out
				}
			}
		}
	}

	return Fingerprint(strconv.FormatInt(totalSize, 16)), nil
}

func stats(path string) []os.FileInfo {
	file, err := os.Open(path)
	if err != nil {
		return []os.FileInfo{} // ignore the error
	}
	defer file.Close()

	stats, err := file.Readdir(0)
	if err != nil {
		return []os.FileInfo{} // ignore the error
	}

	return stats
}

func calculateSparseFingerprint(path string, fileSize int64, h hash.Hash) (Fingerprint, error) {
	buffer := make([]byte, sparseFingerprintSize)

	file, err := os.Open(path)
	if err != nil {
		return Empty, err
	}
	defer file.Close()

	// start
	count, err := file.Read(buffer)
	if err != nil {
		return Empty, err
	}
	h.Write(buffer[:count])

	// middle
	_, err = file.Seek((fileSize-sparseFingerprintSize)/2, 0)
	if err != nil {
		return Empty, err
	}

	count, err = file.Read(buffer)
	if err != nil {
		return Empty, err
	}
	h.Write(buffer[:count])

	// end
	_, err = file.Seek(-sparseFingerprintSize, 2)
	if err != nil {
		return Empty, err
	}

	count, err = file.Read(buffer)
	if err != nil {
		return Empty, err
	}
	h.Write(buffer[:count])

	sum := h.Sum(make([]byte, 0, 64))
	fingerprint := hex.EncodeToString(sum)

	return Fingerprint(fingerprint), nil
}

func calculateRegularFingerprint(path string, h hash.Hash) (Fingerprint, error) {
	file, err := os.Open(path)
	if err != nil {
		return Empty, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for count := 0; err == nil; count, err = file.Read(buffer) {
		h.Write(buffer[:count])
	}

	sum := h.Sum(make([]byte, 0, 64))
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
