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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"strings"
)

const sparseFingerprintThreshold = 5 * 1024 * 1024
const sparseFingerprintSize = 512 * 1024

// Create a fingerprint using the specified algorithm.
func Create(path, fingerprintAlgorithm string) (Fingerprint, error) {
	switch fingerprintAlgorithm {
	case "dynamic:SHA256", "":
		return dynamicFingerprint(path, sha256.New())
	case "dynamic:SHA1":
		return dynamicFingerprint(path, sha1.New())
	case "dynamic:MD5":
		return dynamicFingerprint(path, md5.New())
	case "SHA256":
		return regularFingerprint(path, sha256.New())
	case "SHA1":
		return regularFingerprint(path, sha1.New())
	case "MD5":
		return regularFingerprint(path, md5.New())
	case "symlinkTargetName":
		return symlinkTargetName(path, true)
	case "symlinkTargetNameNoExt":
		return symlinkTargetName(path, false)
	default:
		return "", fmt.Errorf("unsupported fingerprint algorithm '%v'.", fingerprintAlgorithm)
	}
}

// unexported

func regularFingerprint(path string, h hash.Hash) (Fingerprint, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return EMPTY, fmt.Errorf("'%v': could not determine if path is a directory: %v", path, err)
	}
	if stat.IsDir() {
		return EMPTY, nil
	}

	return calculateRegularFingerprint(path, h)
}

func dynamicFingerprint(path string, h hash.Hash) (Fingerprint, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return EMPTY, fmt.Errorf("'%v': could not determine if path is a directory: %v", path, err)
	}
	if stat.IsDir() {
		return EMPTY, nil
	}

	fileSize := stat.Size()

	if fileSize > sparseFingerprintThreshold {
		return calculateSparseFingerprint(path, fileSize, h)
	}

	return calculateRegularFingerprint(path, h)
}

// Uses the symoblic target's filename as the fingerprint
func symlinkTargetName(path string, includeExtension bool) (Fingerprint, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return EMPTY, fmt.Errorf("'%v': could not determine if path is symbolic link: %v", path, err)
	}

	if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		// for other files, use the base name
		return Fingerprint(stat.Name()), nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("'%v': could not determine targe of symbolic link: %v", path, err)
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

func calculateSparseFingerprint(path string, fileSize int64, h hash.Hash) (Fingerprint, error) {
	buffer := make([]byte, sparseFingerprintSize)
	hash := sha256.New()

	file, err := os.Open(path)
	if err != nil {
		return EMPTY, err
	}
	defer file.Close()

	// start
	count, err := file.Read(buffer)
	if err != nil {
		return EMPTY, err
	}
	hash.Write(buffer[:count])

	// middle
	_, err = file.Seek((fileSize-sparseFingerprintSize)/2, 0)
	if err != nil {
		return EMPTY, err
	}

	count, err = file.Read(buffer)
	if err != nil {
		return EMPTY, err
	}
	hash.Write(buffer[:count])

	// end
	_, err = file.Seek(-sparseFingerprintSize, 2)
	if err != nil {
		return EMPTY, err
	}

	count, err = file.Read(buffer)
	if err != nil {
		return EMPTY, err
	}
	hash.Write(buffer[:count])

	sum := hash.Sum(make([]byte, 0, 64))
	fingerprint := hex.EncodeToString(sum)

	return Fingerprint(fingerprint), nil
}

func calculateRegularFingerprint(path string, h hash.Hash) (Fingerprint, error) {
	file, err := os.Open(path)
	if err != nil {
		return EMPTY, err
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
