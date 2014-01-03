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

package entities

type Tag struct {
	Id   uint
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

func (tags Tags) ContainsName(name string) bool {
	for _, tag := range tags {
		if tag.Name == name {
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
	Id        uint
	Name      string
	FileCount uint
}
