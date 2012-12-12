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

type RenameCommand struct{}

func (RenameCommand) Name() string {
	return "rename"
}

func (RenameCommand) Synopsis() string {
	return "Rename a tag"
}

func (RenameCommand) Description() string {
	return `tmsu rename OLD NEW

    Renames a tag from OLD to NEW.

    Attempting to rename a tag with a new name for which a tag already exists will result in an error.
    To merge tags use the 'merge' command instead.`
}

func (RenameCommand) Options() []cli.Option {
	return []cli.Option{}
}

func (RenameCommand) Exec(args []string) error {
	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, err := store.TagByName(sourceTagName)
	if err != nil {
		return err
	}
	if sourceTag == nil {
		return errors.New("No such tag '" + sourceTagName + "'.")
	}

	err = cli.ValidateTagName(destTagName)
	if err != nil {
		return err
	}

	destTag, err := store.TagByName(destTagName)
	if err != nil {
		return err
	}
	if destTag != nil {
		return errors.New("A tag with name '" + destTagName + "' already exists.")
	}

	_, err = store.RenameTag(sourceTag.Id, destTagName)
	if err != nil {
		return err
	}

	return nil
}
