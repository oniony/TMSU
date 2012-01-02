/*
Copyright 2011 Paul Ruane.

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

package main

import (
	"errors"
)

type MergeCommand struct{}

func (MergeCommand) Name() string {
	return "merge"
}

func (MergeCommand) Summary() string {
	return "merges two tags together"
}

func (MergeCommand) Help() string {
	return `  tmsu merge SRC DEST
        
Merges SRC into DEST resulting in a single tag of name DEST.`
}

func (MergeCommand) Exec(args []string) error {
	db, error := OpenDatabase(databasePath())
	if error != nil { return error }
	//defer db.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, error := db.TagByName(sourceTagName)
	if error != nil { return error }
	if sourceTag == nil { return errors.New("No such tag '" + sourceTagName + "'.") }

	destTag, error := db.TagByName(destTagName)
	if error != nil { return error }
	if destTag == nil { return errors.New("No such tag '" + destTagName + "'.") }

    fileTags, error := db.FileTagsByTagId(sourceTag.Id)
    if error != nil { return error }

    for _, fileTag := range fileTags {
        destFileTag, error := db.FileTagByFileIdAndTagId(fileTag.FileId, destTag.Id)
        if error != nil { return error }
        if destFileTag != nil { continue }

        _, error = db.AddFileTag(fileTag.FileId, destTag.Id)
        if error != nil { return error }
    }

    error = db.RemoveFileTagsByTagId(sourceTag.Id)
    if error != nil { return error }

	error = db.DeleteTag(sourceTag.Id)
	if error != nil { return error }

	error = db.Close()
	if error != nil { return error }

	return nil
}
