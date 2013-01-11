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

package commands

import (
	"errors"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type CopyCommand struct {
	verbose bool
}

func (CopyCommand) Name() cli.CommandName {
	return "copy"
}

func (CopyCommand) Synopsis() string {
	return "Create a copy of a tag"
}

func (CopyCommand) Description() string {
	return `tmsu copy TAG NEW

Creates a new tag NEW applied to the same set of files as TAG.`
}

func (CopyCommand) Options() cli.Options {
	return cli.Options{}
}

func (command CopyCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	store, err := storage.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	sourceTagName := args[0]
	destTagName := args[1]

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
	if destTag != nil {
		return errors.New("A tag with name '" + destTagName + "' already exists.")
	}

	if command.verbose {
		log.Infof("copying tag '%v' to '%v'.", sourceTagName, destTagName)
	}

	_, err = store.CopyTag(sourceTag.Id, destTagName)
	if err != nil {
		return err
	}

	return nil
}
