package main

import (
	"errors"
)

type MergeCommand struct{}

func (this MergeCommand) Name() string {
	return "merge"
}

func (this MergeCommand) Summary() string {
	return "merges two tags together"
}

func (this MergeCommand) Help() string {
	return `  tmsu merge SRC DEST
        
Merges SRC into DEST resulting in a single tag of name DEST.`
}

func (this MergeCommand) Exec(args []string) error {
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
	if destTag == nil {
		return errors.New("No such tag '" + destTagName + "'.")
	}

	error = db.MigrateFileTags(sourceTag.Id, destTag.Id)
	if error != nil {
		return error
	}

	error = db.DeleteTag(sourceTag.Id)
	if error != nil {
		return error
	}

	return nil
}
