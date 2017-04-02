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

package database

import (
	"fmt"
	"github.com/oniony/TMSU/entities"
)

type DatabaseNotFoundError struct {
	Path string
}

func (err DatabaseNotFoundError) Error() string {
	return fmt.Sprintf("no database at '%v'", err.Path)
}

type DatabaseAccessError struct {
	DatabasePath string
	Reason       error
}

func (err DatabaseAccessError) Error() string {
	return fmt.Sprintf("cannot access database at '%v': %v", err.DatabasePath, err.Reason)
}

type DatabaseTransactionError struct {
	DatabasePath string
	Reason       error
}

func (err DatabaseTransactionError) Error() string {
	return fmt.Sprintf("database transaction error: %v", err.Reason)
}

type DatabaseQueryError struct {
	DatabasePath string
	Query        string
	Reason       error
}

func (err DatabaseQueryError) Error() string {
	return fmt.Sprintf("database query failed: %v", err.Reason)
}

type NoSuchFileError struct {
	FileId entities.FileId
}

func (err NoSuchFileError) Error() string {
	return fmt.Sprintf("no such file #%v", err.FileId)
}

type NoSuchValueError struct {
	ValueId entities.ValueId
}

func (err NoSuchValueError) Error() string {
	return fmt.Sprintf("no such value #%v", err.ValueId)
}

type NoSuchQueryError struct {
	Query string
}

func (err NoSuchQueryError) Error() string {
	return fmt.Sprintf("no such query '%v'", err.Query)
}

type NoSuchFileTagError struct {
	FileId  entities.FileId
	TagId   entities.TagId
	ValueId entities.ValueId
}

func (err NoSuchFileTagError) Error() string {
	return fmt.Sprintf("no such file-tag for file #%v, tag #%v and value #%v.", err.FileId, err.TagId, err.ValueId)
}

type NoSuchImplicationError struct {
	TagValuePair        entities.TagIdValueIdPair
	ImpliedTagValuePair entities.TagIdValueIdPair
}

func (err NoSuchImplicationError) Error() string {
	return fmt.Sprintf("no such implication where #%v implies #%v", err.TagValuePair, err.ImpliedTagValuePair)
}

type NoSuchSettingError struct {
	Name string
}

func (err NoSuchSettingError) Error() string {
	return fmt.Sprintf("no such setting '%v'", err.Name)
}
