/*
Copyright 2011-2012 Paul Ruane.

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

package commands

import (
	"errors"
	"tmsu/cli"
	"tmsu/storage"
)

type MergeCommand struct{}

func (MergeCommand) Name() string {
	return "merge"
}

func (MergeCommand) Synopsis() string {
	return "Merge tags"
}

func (MergeCommand) Description() string {
	return `tmsu merge TAG... DEST
        
Merges TAGs into tag DEST resulting in a single tag of name DEST.`
}

func (MergeCommand) Options() []cli.Option {
	return []cli.Option{}
}

func (MergeCommand) Exec(args []string) error {
	if len(args) < 2 {
		return errors.New("Too few arguments.")
	}

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	destTagName := args[len(args)-1]

	for _, sourceTagName := range args[0 : len(args)-1] {
		sourceTag, err := store.Db.TagByName(sourceTagName)
		if err != nil {
			return err
		}
		if sourceTag == nil {
			return errors.New("No such tag '" + sourceTagName + "'.")
		}

		destTag, err := store.Db.TagByName(destTagName)
		if err != nil {
			return err
		}
		if destTag == nil {
			return errors.New("No such tag '" + destTagName + "'.")
		}

		fileTags, err := store.Db.FileTagsByTagId(sourceTag.Id)
		if err != nil {
			return err
		}

		for _, fileTag := range fileTags {
			destFileTag, err := store.Db.FileTagByFileIdAndTagId(fileTag.FileId, destTag.Id)
			if err != nil {
				return err
			}
			if destFileTag != nil {
				continue
			}

			_, err = store.Db.AddFileTag(fileTag.FileId, destTag.Id)
			if err != nil {
				return err
			}
		}

		err = store.Db.RemoveFileTagsByTagId(sourceTag.Id)
		if err != nil {
			return err
		}

		err = store.Db.DeleteTag(sourceTag.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
