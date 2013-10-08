/*
Copyright 2011-2013 Paul Ruane.

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

package query

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

var tagChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct}

type Token interface {
}

func Type(token Token) string {
	switch token.(type) {
	case TagToken:
		return "tag"
	case OpenParenToken:
		return "("
	case CloseParenToken:
		return ")"
	case NotOperatorToken:
		return "not"
	case AndOperatorToken:
		return "and"
	case OrOperatorToken:
		return "or"
	case EndToken:
		return "EOF"
	case nil:
		return "nil"
	default:
		return "unkown"
	}
}

type EndToken struct {
}

type OpenParenToken struct {
}

type CloseParenToken struct {
}

type TagToken struct {
	name string
}

type NotOperatorToken struct {
}

type AndOperatorToken struct {
}

type OrOperatorToken struct {
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
			return nil, err
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
		return nil, err
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
	case unicode.IsOneOf(tagChars, r):
		return scanner.readTagTokenOrOperatorToken(r)
	default:
		return nil, fmt.Errorf("Unepxected character '%v'.", r)
	}

	panic("unreachable")
}

func (scanner *Scanner) readTagTokenOrOperatorToken(r rune) (Token, error) {
	text, err := scanner.readString(r)
	if err != nil {
		return nil, err
	}

	switch text {
	case "not":
		return NotOperatorToken{}, nil
	case "and":
		return AndOperatorToken{}, nil
	case "or":
		return OrOperatorToken{}, nil
	default:
		return TagToken{text}, nil
	}
}

func (scanner *Scanner) readString(r rune) (string, error) {
	text := string(r)

	stop := false

	for !stop {
		r, _, err := scanner.stream.ReadRune()

		if err == io.EOF {
			return text, nil
		}
		if err != nil {
			return "", err
		}

		switch {
		case unicode.IsSpace(r), r == rune(')'), r == rune('('):
			scanner.stream.UnreadRune()
			return text, nil
		case unicode.IsOneOf(tagChars, r):
			text += string(r)
		default:
			return "", fmt.Errorf("Unexpected character '%v'.", r)
		}
	}

	panic("unreachable")
}
