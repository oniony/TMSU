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

package entities

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type TagId uint

type TagIds []TagId

func (tagIds TagIds) Len() int {
	return len(tagIds)
}

func (tagIds TagIds) Less(i, j int) bool {
	return tagIds[i] < tagIds[j]
}

func (tagIds TagIds) Swap(i, j int) {
	tagIds[i], tagIds[j] = tagIds[j], tagIds[i]
}

func (tagIds TagIds) Uniq() TagIds {
	if len(tagIds) == 0 {
		return tagIds
	}

	sort.Sort(tagIds)
	uniq := TagIds{tagIds[0]}
	previous := tagIds[0]

	for _, tagId := range tagIds[1:] {
		if tagId != previous {
			uniq = append(uniq, tagId)
			previous = tagId
		}
	}

	return uniq
}

type Tag struct {
	Id   TagId
	Name string
}

type Tags []*Tag

func (tags Tags) Len() int {
	return len(tags)
}

func (tags Tags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}

func (tags Tags) Less(i, j int) bool {
	return tags[i].Name < tags[j].Name
}

func (tags Tags) Contains(searchTag *Tag) bool {
	for _, tag := range tags {
		if tag.Id == searchTag.Id {
			return true
		}
	}

	return false
}

func (tags Tags) ContainsCasedName(name string, ignoreCase bool) bool {
	if ignoreCase {
		name = strings.ToLower(name)
	}

	for _, tag := range tags {
		tagName := tag.Name
		if ignoreCase {
			tagName = strings.ToLower(tagName)
		}

		if tagName == name {
			return true
		}
	}

	return false
}

func (tags Tags) Any(predicate func(*Tag) bool) bool {
	for _, tag := range tags {
		if predicate(tag) {
			return true
		}
	}

	return false
}

type TagFileCount struct {
	Id        TagId
	Name      string
	FileCount uint
}

func ValidateTagName(tagName string) error {
	switch tagName {
	case "":
		return fmt.Errorf("tag name cannot be empty")
	case ".", "..":
		return fmt.Errorf("tag name cannot be '.' or '..'") // cannot be used in the VFS
	case "and", "AND", "or", "OR", "not", "NOT":
		return fmt.Errorf("tag name cannot be a logical operator: 'and', 'or' or 'not'") // used in query language
	case "eq", "EQ", "ne", "NE", "lt", "LT", "gt", "GT", "le", "LE", "ge", "GE":
		return fmt.Errorf("tag name cannot be a comparison operator: 'eq', 'ne', 'gt', 'lt', 'ge' or 'le'") // used in query language
	}

	for _, ch := range tagName {
		if !unicode.IsOneOf(validTagChars, ch) {
			if unicode.IsPrint(ch) {
				return fmt.Errorf("tag names cannot contain '%c'", ch)
			}

			return fmt.Errorf("tag names cannot contain %U", ch)
		}
	}

	return nil
}

// unexported

var validTagChars = []*unicode.RangeTable{unicode.Letter, unicode.Number, unicode.Punct, unicode.Symbol, unicode.Space}
