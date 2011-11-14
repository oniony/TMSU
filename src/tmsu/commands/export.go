package main

import (
	"errors"
	"fmt"
	"strings"
)

type ExportCommand struct{}

func (this ExportCommand) Name() string {
	return "export"
}

func (this ExportCommand) Summary() string {
	return "exports the tag database"
}

func (this ExportCommand) Help() string {
	return `tmsu export
        
dumps the tag database to standard output as comma-separated values (CSV)`
}

func (this ExportCommand) Exec(args []string) error {
	if len(args) != 0 {
		return errors.New("Unpected argument to command '" + this.Name() + "'.")
	}

	db, error := OpenDatabase(databasePath())
	if error != nil {
		return error
	}
	defer db.Close()

	files, error := db.Files()
	if error != nil {
		return error
	}

	for _, file := range *files {
		fmt.Printf("%v,", file.Path)

		tags, error := db.TagsByFileId(file.Id)
		if error != nil {
			return error
		}

		tagNames := make([]string, 0, len(*tags))
		for _, tag := range *tags {
			tagNames = append(tagNames, tag.Name)
		}
		fmt.Println(strings.Join(tagNames, ","))
	}

	return nil
}
