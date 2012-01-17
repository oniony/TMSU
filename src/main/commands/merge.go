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
	return `tmsu merge SRC DEST
        
Merges SRC into DEST resulting in a single tag of name DEST.`
}

func (MergeCommand) Exec(args []string) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	//defer db.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, err := db.TagByName(sourceTagName)
	if err != nil {
		return err
	}
	if sourceTag == nil {
		return errors.New("No such tag '" + sourceTagName + "'.")
	}

	destTag, err := db.TagByName(destTagName)
	if err != nil {
		return err
	}
	if destTag == nil {
		return errors.New("No such tag '" + destTagName + "'.")
	}

	fileTags, err := db.FileTagsByTagId(sourceTag.Id)
	if err != nil {
		return err
	}

	for _, fileTag := range fileTags {
		destFileTag, err := db.FileTagByFileIdAndTagId(fileTag.FileId, destTag.Id)
		if err != nil {
			return err
		}
		if destFileTag != nil {
			continue
		}

		_, err = db.AddFileTag(fileTag.FileId, destTag.Id)
		if err != nil {
			return err
		}
	}

	err = db.RemoveFileTagsByTagId(sourceTag.Id)
	if err != nil {
		return err
	}

	err = db.DeleteTag(sourceTag.Id)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}
