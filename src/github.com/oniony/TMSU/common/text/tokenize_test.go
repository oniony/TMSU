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

package text

import (
	"testing"
)

func TestSimple(test *testing.T) {
	words := Tokenize("one two three")

	if len(words) != 3 || words[0] != "one" || words[1] != "two" || words[2] != "three" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestQuoted(test *testing.T) {
	words := Tokenize(`one 'two three' four`)

	if len(words) != 3 || words[0] != "one" || words[1] != "two three" || words[2] != "four" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestDoubleQuoted(test *testing.T) {
	words := Tokenize(`one "two three" four`)

	if len(words) != 3 || words[0] != "one" || words[1] != "two three" || words[2] != "four" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestSingleQuoteInsideDoubleQuoted(test *testing.T) {
	words := Tokenize(`one "two 'three'" four`)

	if len(words) != 3 || words[0] != "one" || words[1] != "two 'three'" || words[2] != "four" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestDoubleQuoteInsideSingleQuoted(test *testing.T) {
	words := Tokenize(`one 'two "three"' four`)

	if len(words) != 3 || words[0] != "one" || words[1] != `two "three"` || words[2] != "four" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestEscapedSingleQuote(test *testing.T) {
	words := Tokenize(`one\'s 'two\'s three\'s' four\'s`)

	if len(words) != 3 || words[0] != "one's" || words[1] != "two's three's" || words[2] != "four's" {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestEscapedDoubleQuote(test *testing.T) {
	words := Tokenize(`one\"s "two\"s three\"s" four\"s`)

	if len(words) != 3 || words[0] != `one"s` || words[1] != `two"s three"s` || words[2] != `four"s` {
		test.Fatalf("tokenization failed: %v", words)
	}
}

func TestComplex(test *testing.T) {
	words := Tokenize(`'one' "two three" four 'five "six" seven' "eight"`)

	if len(words) != 5 || words[0] != "one" || words[1] != "two three" || words[2] != "four" || words[3] != `five "six" seven` || words[4] != "eight" {
		test.Fatalf("tokenization failed: %v", words)
	}
}
