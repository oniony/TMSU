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
	"fmt"
	"os"
	"path/filepath"
)

type TagsCommand struct{}

func (TagsCommand) Name() string {
	return "tags"
}

func (TagsCommand) Synopsis() string {
	return "list tags"
}

func (TagsCommand) Description() string {
	return `tmsu tags [FILE]...
tmsu tags --list

Lists the tags applied to FILEs (files in the current directory by default).

  --list    lists all tags defined`
}

func (command TagsCommand) Exec(args []string) error {
	argCount := len(args)

	if argCount == 1 && args[0] == "--list" {
		return command.listAllTags()
	}

	return command.listTags(args)
}

func (TagsCommand) listAllTags() error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	tags, err := db.Tags()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTags(paths []string) error {
	db, err := OpenDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	switch len(paths) {
	case 0:
		return command.listTagsRecursive(db, []string{"."})
	case 1:
		return command.listTagsForPath(db, paths[0])
	default:
		return command.listTagsRecursive(db, paths)
	}

	return command.listTagsRecursive(db, paths)
}

func (command TagsCommand) listTagsForPath(db *Database, path string) error {
	tags, err := command.tagsForPath(db, path)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	return nil
}

func (command TagsCommand) listTagsRecursive(db *Database, paths []string) error {
	for _, path := range paths {
		fileInfo, err := os.Lstat(path)
		if err != nil {
			return err
		}

        tags, err := command.tagsForPath(db, path)
        if err != nil {
            return err
        }

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

		if fileInfo.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				switch terr := err.(type) {
				case *os.PathError:
					switch terr.Err {
					case os.EACCES:
						warn("'%v': permission denied.", path)
					default:
						warn("'%v': %v", path, err)
					}
				default:
					warn("'%v': %v", path, err)
				}
				continue
			}
			defer file.Close()

			dirNames, err := file.Readdirnames(0)
			if err != nil {
				return err
			}

			childPaths := make([]string, len(dirNames))
			for index, dirName := range dirNames {
				childPaths[index] = filepath.Join(path, dirName)
			}

			err = command.listTagsRecursive(db, childPaths)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (TagsCommand) tagsForPath(db *Database, path string) ([]Tag, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := db.FileByPath(absPath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, nil
	}

	tags, err := db.TagsByFileId(file.Id)
	if err != nil {
		return nil, err
	}

	return tags, err
}
