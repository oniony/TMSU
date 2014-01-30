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
	"fmt"
	"unicode"
)

var blankError = errors.New("")

var validChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol}

type TagValuePair struct {
	TagId   uint
	ValueId uint
}

func ValidateTagNames(tagNames []string) error {
	for _, tagName := range tagNames {
		if err := ValidateTagName(tagName); err != nil {
			return err
		}
	}

	return nil
}

func ValidateTagName(tagName string) error {
	switch tagName {
	case "":
		return errors.New("tag name cannot be empty.")
	case ".", "..":
		return errors.New("tag name cannot be '.' or '..'.") // cannot be used in the VFS
	case "and", "or", "not", "AND", "OR", "NOT":
		return errors.New("tag name cannot be a logical operator, 'and', 'or' or 'not'.") // used in query language
	}

	if tagName[0] == '-' {
		return errors.New("tag name cannot start with a minus: '-'.") // used in query language
	}

	for _, r := range tagName {
		switch r {
		case '(', ')':
			return errors.New("tag names cannot contain parentheses, '(' or ')'.") // used in query language
		case ',':
			return errors.New("tag names cannot contain commas, ','.") // reserved for tag delimiter
		case '=':
			return errors.New("tag names cannot contain comparison operators, '=', '<' or '>'.") // reserved for tag values
		case ' ', '\t':
			return errors.New("tag names cannot contain spaces or tabs.") // used as tag delimiter
		case '/':
			return errors.New("tag names cannot contain slashes, '/'.") // cannot be used in the VFS
		}

		if !unicode.IsOneOf(validChars, r) {
			return fmt.Errorf("tag names cannot contain '%v'.", r)
		}
	}

	return nil
}
