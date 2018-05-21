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
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultGeneration(test *testing.T) {
	testCreateForSmallFile(test, "", "cdf701ac9e4258a8efec453930c73d698d12d7e83c38a049a1f1a64375fbf776")
	testCreateForLargeFile(test, "", "0a9f9c7cd5939b04ad4bb7d14f801fe671c1b622d0e3b7769798b14dbdbf07f1")
}

func TestMD5Generation(test *testing.T) {
	testCreateForSmallFile(test, "MD5", "a758071b3c2fe43c9a9b91db5077cd12")
	testCreateForLargeFile(test, "MD5", "40cb0a2f629169e30c2ef707255b33d5")
}

func TestSHA1Generation(test *testing.T) {
	testCreateForSmallFile(test, "SHA1", "09bc65c6f6588b802177632a81b3afbe3358b7f3")
	testCreateForLargeFile(test, "SHA1", "6dfb0884a5f2738700b9beb5473f3dd9c9fd2762")
}

func TestSHA256Generation(test *testing.T) {
	testCreateForSmallFile(test, "SHA256", "cdf701ac9e4258a8efec453930c73d698d12d7e83c38a049a1f1a64375fbf776")
	testCreateForLargeFile(test, "SHA256", "a4bd6407e40326c126f10412e245e4491c511636dbeddc3d2b16b41700017bc9")
}

func TestBLAKE2bGeneration(test *testing.T) {
	testCreateForSmallFile(test, "BLAKE2b", "76b01099c5121e2436f3cb201f3917e4f46eae7dac8ac0c941b1729101e91de4")
	testCreateForLargeFile(test, "BLAKE2b", "fdc4dc9cebbd6f162b3dad4d196646df430dbae8c547df01447285da55247087")
}

func TestDynamicMD5Generation(test *testing.T) {
	testCreateForSmallFile(test, "dynamic:MD5", "a758071b3c2fe43c9a9b91db5077cd12")
	testCreateForLargeFile(test, "dynamic:MD5", "668a4b622482b9fd30b1ad0eac4ab8f1")
}

func TestDynamicSHA1Generation(test *testing.T) {
	testCreateForSmallFile(test, "dynamic:SHA1", "09bc65c6f6588b802177632a81b3afbe3358b7f3")
	testCreateForLargeFile(test, "dynamic:SHA1", "30af88de9e731520fbcb4ec5f7276af8e06eb61b")
}

func TestDynamicSHA256Generation(test *testing.T) {
	testCreateForSmallFile(test, "dynamic:SHA256", "cdf701ac9e4258a8efec453930c73d698d12d7e83c38a049a1f1a64375fbf776")
	testCreateForLargeFile(test, "dynamic:SHA256", "0a9f9c7cd5939b04ad4bb7d14f801fe671c1b622d0e3b7769798b14dbdbf07f1")
}

func TestDynamicBLAKE2bGeneration(test *testing.T) {
	testCreateForSmallFile(test, "dynamic:BLAKE2b", "76b01099c5121e2436f3cb201f3917e4f46eae7dac8ac0c941b1729101e91de4")
	testCreateForLargeFile(test, "dynamic:BLAKE2b", "137c5b1e9e8107c176de7fb7a38f7670bb31364fadb2b5b883737c8732c78327")
}

func TestNoneGeneration(test *testing.T) {
	testCreateForSmallFile(test, "none", "")
	testCreateForLargeFile(test, "none", "")
}

// unexported

func testCreateForSmallFile(test *testing.T, algorithm string, expectedFingerprint Fingerprint) {
	testCreateForFile(test, algorithm, 2*1024*1024, expectedFingerprint)
}

func testCreateForLargeFile(test *testing.T, algorithm string, expectedFingerprint Fingerprint) {
	testCreateForFile(test, algorithm, 6*1024*1024, expectedFingerprint)
}

func testCreateForFile(test *testing.T, algorithm string, size uint, expectedFingerprint Fingerprint) {
	tempFilePath := filepath.Join(os.TempDir(), "tmsu-fingerprint")
	file, err := os.Create(tempFilePath)
	if err != nil {
		test.Fatal(err.Error())
	}
	defer os.Remove(tempFilePath)

	_, err = file.WriteAt([]byte("!"), int64(size-1))
	if err != nil {
		test.Fatal(err.Error())
	}

	fingerprint, err := Create(tempFilePath, algorithm, "none", "none")
	if err != nil {
		test.Fatal(err.Error())
	}

	if fingerprint != expectedFingerprint {
		test.Fatalf("Fingerprint incorrect: expected '%v' but was '%v'", expectedFingerprint, fingerprint)
	}
}
