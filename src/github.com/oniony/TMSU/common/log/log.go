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

package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var Verbosity uint = 1

func Fatal(values ...interface{}) {
	log(os.Stderr, values...)
	os.Exit(1)
}

func Fatalf(format string, values ...interface{}) {
	logf(os.Stderr, format, values...)
	os.Exit(1)
}

func Warn(values ...interface{}) {
	log(os.Stderr, values...)
}

func Warnf(format string, values ...interface{}) {
	logf(os.Stderr, format, values...)
}

func Info(verbosity uint, values ...interface{}) {
	if verbosity > Verbosity {
		return
	}

	log(os.Stdout, values...)
}

func Infof(verbosity uint, format string, values ...interface{}) {
	if verbosity > Verbosity {
		return
	}

	logf(os.Stdout, format, values...)
}

// unexported

func log(dest io.Writer, values ...interface{}) {
	if Verbosity > 1 {
		fmt.Fprintf(dest, "%v: ", time.Now())
	}

	fmt.Fprintf(dest, "tmsu: ")
	fmt.Fprintln(dest, values...)
}

func logf(dest io.Writer, format string, values ...interface{}) {
	if Verbosity > 1 {
		fmt.Printf("%v: ", time.Now())
	}

	format = "tmsu: " + format + "\n"
	fmt.Fprintf(dest, format, values...)
}
