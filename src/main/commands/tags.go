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
package main
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type TagsCommand struct{}

func (this TagsCommand) Name() string {
	return "tags"
}

func (this TagsCommand) Summary() string {
	return "lists all tags or tags applied to a file or files"
}

func (this TagsCommand) Help() string {
	return `  tmsu tags --all
  tmsu tags [FILE]...

Lists the tags applied to FILEs (the current directory by default).

  --all    show the complete set of tags`
}

func (this TagsCommand) Exec(args []string) error {
    argCount := len(args)

    if argCount == 0 {
        return this.listTags(".")
    } else if argCount == 1 && args[0] == "--all" {
        return this.listAllTags()
    } else {
	    this.listTags(args...)
	}

	return nil
}

func (this TagsCommand) listAllTags() error {
	db, error := OpenDatabase(databasePath())
	if error != nil { return error }
	defer db.Close()

	tags, error := db.Tags()
	if error != nil { return error }

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (this TagsCommand) listTags(paths ...string) error {
	db, error := OpenDatabase(databasePath())
	if error != nil { return error }
	defer db.Close()

    if len(paths) == 1 {
        fileInfo, error := os.Lstat(paths[0])
        if error != nil { return error }

        if fileInfo.Mode() & os.ModeType == 0 {
            tags, error := this.tagsForPath(db, paths[0])
            if error != nil { return error }
            if tags == nil { return nil }

            for _, tag := range tags {
                fmt.Println(tag.Name)
            }

            return nil
        }
    }

    return this.listTagsRecursive(db, paths)
}

func (this TagsCommand) listTagsRecursive(db *Database, paths []string) error {
    for _, path := range paths {
        fileInfo, error := os.Lstat(path)
        if error != nil { return error }

        if fileInfo.Mode() & os.ModeType == 0 {
            tags, error := this.tagsForPath(db, path)
            if error != nil { return error }
            if tags == nil { continue }

            if len(tags) > 0 {
                fmt.Printf("%v: ", path)

                for index, tag := range tags {
                    if index > 0 {
                        fmt.Print(" ")
                    }

                    fmt.Print(tag.Name)
                }

                fmt.Println()
            }
        } else if fileInfo.IsDir() {
            file, error := os.Open(path)
            if error != nil { return error }
            defer file.Close()

            dirNames, error := file.Readdirnames(0)
            if error != nil { return error }

            childPaths := make([]string, len(dirNames))
            for index, dirName := range dirNames {
                childPaths[index] = filepath.Join(path, dirName)
            }

            error = this.listTagsRecursive(db, childPaths)
            if error != nil { return error }
        }
    }

    return nil
}

func (this TagsCommand) tagsForPath(db *Database, path string) ([]Tag, error) {
	absPath, error := filepath.Abs(path)
	if error != nil { return nil, error }

	file, error := db.FileByPath(absPath)
	if error != nil { return nil, error }
	if file == nil { return nil, nil }

	tags, error := db.TagsByFileId(file.Id)
	if error != nil { return nil, error }

	return tags, error
}
