package main

import (
	"errors"
	"fmt"
)

type RenameCommand struct{}

func (this RenameCommand) Name() string {
	return "rename"
}

func (this RenameCommand) Summary() string {
	return "renames a tag"
}

func (this RenameCommand) Help() string {
	return `  tmsu rename OLD NEW

Renames a tag from OLD to NEW.

Attempting to rename a tag with a new name for which a tag already exists will result in an error.
To merge tags use the 'merge' command instead.`
}

func (this RenameCommand) Exec(args []string) error {
	db, error := OpenDatabase(databasePath())
	if error != nil {
		return error
	}
	defer db.Close()

	sourceTagName := args[0]
	destTagName := args[1]

	sourceTag, error := db.TagByName(sourceTagName)
	if error != nil {
		return error
	}
	if sourceTag == nil {
		return errors.New("No such tag '" + sourceTagName + "'.")
	}

	destTag, error := db.TagByName(destTagName)
	if error != nil {
		return error
	}
	if destTag != nil {
		return errors.New("A tag with name '" + destTagName + "' already exists.")
	}

	_, error = db.RenameTag(sourceTag.Id, destTagName)
	if error != nil {
		return error
	}

	return nil
}
