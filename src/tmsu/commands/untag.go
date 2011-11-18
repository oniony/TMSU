// Copyright 2011 Paul Ruane. All rights reserved.

package main

import (
	"errors"
	"path/filepath"
)

type UntagCommand struct{}

func (this UntagCommand) Name() string {
	return "untag"
}

func (this UntagCommand) Summary() string {
	return "removes all tags or specific tags from a file"
}

func (this UntagCommand) Help() string {
	return `  tmsu untag FILE TAG...
  tmsu untag FILE

Disassociates the specified file FILE with the tag(s) specified.

If no tags are specified then the file will be stripped of all tags.`
}

func (this UntagCommand) Exec(args []string) error {
	if len(args) < 2 {
		return errors.New("File to untag and tags to remove must be specified.")
	}

	error := this.untagPath(args[0], args[1:])
	if error != nil {
		return error
	}

	return nil
}

// implementation

func (this UntagCommand) untagPath(path string, tagNames []string) error {
	absPath, error := filepath.Abs(path)
	if error != nil {
		return error
	}

	db, error := OpenDatabase(databasePath())
	if error != nil {
		return error
	}
	defer db.Close()

	file, error := db.FileByPath(absPath)
	if error != nil {
		return error
	}
	if file == nil {
		return errors.New("File '" + path + "' is not tagged.")
	}

	for _, tagName := range tagNames {
		error = this.unapplyTag(db, path, file.Id, tagName)
		if error != nil {
			return error
		}
	}

	hasTags, error := db.AnyFileTagsForFile(file.Id)
	if error != nil {
		return error
	}

	if !hasTags {
		db.RemoveFile(file.Id)
	}

	return nil
}

func (this UntagCommand) unapplyTag(db *Database, path string, fileId uint, tagName string) error {
	tag, error := db.TagByName(tagName)
	if error != nil {
		return error
	}
	if tag == nil {
		errors.New("No such tag" + tagName)
	}

	fileTag, error := db.FileTagByFileIdAndTagId(fileId, tag.Id)
	if error != nil {
		return error
	}
	if fileTag == nil {
		errors.New("File '" + path + "' is not tagged '" + tagName + "'.")
	}

	if fileTag != nil {
		error := db.RemoveFileTag(fileId, tag.Id)
		if error != nil {
			return error
		}
	} else {
		return errors.New("File '" + path + "' is not tagged '" + tagName + "'.\n")
	}

	return nil
}
