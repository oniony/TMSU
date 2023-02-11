// Copyright 2011-2018 Paul Ruane.

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

package query

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

var symbolChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol}

type Token interface {
}

func Type(token Token) string {
	switch typedToken := token.(type) {
	case SymbolToken:
		return "symbol"
	case OpenParenToken:
		return "'('"
	case CloseParenToken:
		return "')'"
	case NotOperatorToken:
		return "'not'"
	case AndOperatorToken:
		return "'and'"
	case OrOperatorToken:
		return "'or'"
	case ComparisonOperatorToken:
		return typedToken.operator
	case EndToken:
		return "EOF"
	case nil:
		return "nil"
	default:
		return "unknown"
	}
}

type EndToken struct {
}

type OpenParenToken struct {
}

type CloseParenToken struct {
}

type SymbolToken struct {
	name string
}

type NotOperatorToken struct {
}

type AndOperatorToken struct {
}

type OrOperatorToken struct {
}

type ComparisonOperatorToken struct {
	operator string
}

type Scanner struct {
	stream    *strings.Reader
	lookAhead Token
}

func NewScanner(query string) *Scanner {
	return &Scanner{strings.NewReader(query), nil}
}

func (scanner *Scanner) LookAhead() (Token, error) {
	if scanner.lookAhead == nil {
		token, err := scanner.readToken()
		if err != nil {
			return nil, fmt.Errorf("could not look ahead: %v", err)
		}
		scanner.lookAhead = token
	}

	return scanner.lookAhead, nil
}

func (scanner *Scanner) Next() (Token, error) {
	token, err := scanner.LookAhead()
	if err != nil {
		return nil, err
	}

	lookAhead, err := scanner.readToken()
	if err != nil {
		return nil, fmt.Errorf("could not read next token: %v", err)
	}
	scanner.lookAhead = lookAhead

	return token, nil
}

// unexported

func (scanner *Scanner) readToken() (Token, error) {
	r, _, err := scanner.stream.ReadRune()
	for err == nil && unicode.IsSpace(r) {
		r, _, err = scanner.stream.ReadRune()
	}

	if err == io.EOF {
		return EndToken{}, nil
	}
	if err != nil {
		return nil, err
	}

	switch {
	case r == rune('('):
		return OpenParenToken{}, nil
	case r == rune(')'):
		return CloseParenToken{}, nil
	case r == rune('!'), r == rune('='), r == rune('<'), r == rune('>'):
		return scanner.readComparisonOperatorToken(r)
	case unicode.IsOneOf(symbolChars, r), r == rune('\\'):
		scanner.stream.UnreadRune()
		return scanner.readTextToken()
	default:
		return nil, fmt.Errorf("Unepxected character '%v'.", r)
	}
}

func (scanner *Scanner) readTextToken() (Token, error) {
	text, err := scanner.readString()
	if err != nil {
		return nil, err
	}

	switch text {
	case "not", "NOT":
		return NotOperatorToken{}, nil
	case "and", "AND":
		return AndOperatorToken{}, nil
	case "or", "OR":
		return OrOperatorToken{}, nil
	case "eq", "EQ":
		return ComparisonOperatorToken{"="}, nil
	case "ne", "NE":
		return ComparisonOperatorToken{"!="}, nil
	case "lt", "LT":
		return ComparisonOperatorToken{"<"}, nil
	case "gt", "GT":
		return ComparisonOperatorToken{">"}, nil
	case "le", "LE":
		return ComparisonOperatorToken{"<="}, nil
	case "ge", "GE":
		return ComparisonOperatorToken{">="}, nil
	}

	return SymbolToken{text}, nil
}

func (scanner *Scanner) readComparisonOperatorToken(r rune) (Token, error) {
	switch r {
	case rune('='), rune('!'), rune('<'), rune('>'):
		r2, _, err := scanner.stream.ReadRune()
		if err != nil {
			return nil, err
		}

		switch r2 {
		case rune('='):
			return ComparisonOperatorToken{string(r) + "="}, nil
		default:
			scanner.stream.UnreadRune()
			return ComparisonOperatorToken{string(r)}, nil
		}
	default:
		panic("not a valid operator token: " + string(r))
	}
}

func (scanner *Scanner) readString() (string, error) {
	text := ""
	escaped := false
	stop := false

	for !stop {
		r, _, err := scanner.stream.ReadRune()
		if err == io.EOF {
			return text, nil
		}
		if err != nil {
			return "", err
		}

		if escaped {
			text += string(r)
			escaped = false
			continue
		}

		if r == rune('\\') {
			escaped = true
			continue
		}

		switch {
		case unicode.IsSpace(r), r == rune(')'), r == rune('('), r == rune('='), r == rune('!'), r == rune('<'), r == rune('>'):
			scanner.stream.UnreadRune()
			return text, nil
		case unicode.IsOneOf(symbolChars, r):
			text += string(r)
		default:
			return "", fmt.Errorf("Unexpected character '%v'.", r)
		}
	}

	panic("unreachable")
}
