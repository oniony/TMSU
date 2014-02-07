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

package cli

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var blankError = errors.New("")

type TagValuePair struct {
	TagId   uint
	ValueId uint
}

func terminalWidth() int {
	var s winsize

	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&s)))

	return int(s.cols)
}

type winsize struct {
	rows     uint16
	cols     uint16
	pxWidth  uint16
	pxHeight uint16
}
