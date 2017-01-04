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

package terminal

import (
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var getConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

func Colour() bool {
	return false
}

func Width() int {
	outHandle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return 0
	}

	info := &consoleScreenBufferInfo{}
	success, _, _ := syscall.Syscall(getConsoleScreenBufferInfo.Addr(), 2, uintptr(outHandle), uintptr(unsafe.Pointer(info)), 0)
	if int(success) == 0 {
		return 0
	}

	cols := int(info.size.x)
	if cols > 0 {
		cols--
	}

	return cols
}

type (
	short int16
	word  uint16

	coord struct {
		x short
		y short
	}

	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}

	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
)
