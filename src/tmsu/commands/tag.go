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
	"os"
	"path/filepath"
	"strings"
	"tmsu/common"
	"tmsu/database"
	"tmsu/fingerprint"
)

type TagCommand struct{}

func (TagCommand) Name() string {
	return "tag"
}

func (TagCommand) Synopsis() string {
	return "Apply tags to files"
}

func (TagCommand) Description() string {
	return `tmsu tag FILE TAG...
tmsu tag --tags "TAG..." FILE...

Tags the file FILE with the tag(s) specified.

  --tags    allows multiple FILEs to be tagged with the same quoted set of TAGs`
}

func (command TagCommand) Exec(args []string) error {
	if len(args) < 1 {
		return errors.New("Too few arguments.")
	}

	switch args[0] {
	case "--tags":
		if len(args) < 3 {
			return errors.New("Quoted set of tags and at least one file to tag must be specified.")
		}

		tagNames := strings.Fields(args[1])
		paths := args[2:]

		err := command.tagPaths(paths, tagNames)
		if err != nil {
			return err
		}
	default:
		if len(args) < 2 {
			return errors.New("File to tag and tags to apply must be specified.")
		}

		path := args[0]
		tagNames := args[1:]

		err := command.tagPath(path, tagNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPaths(paths []string, tagNames []string) error {
	for _, path := range paths {
		err := command.tagPath(path, tagNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func (command TagCommand) tagPath(path string, tagNames []string) error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := command.addFile(db, absPath)
	if err != nil {
		return err
	}

	for _, tagName := range tagNames {
		_, _, err = command.applyTag(db, path, file.Id, tagName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (TagCommand) applyTag(db *database.Database, path string, fileId uint, tagName string) (*database.Tag, *database.FileTag, error) {
	err := validateTagName(tagName)
	if err != nil {
		return nil, nil, err
	}

	tag, err := db.TagByName(tagName)
	if err != nil {
		return nil, nil, err
	}

	if tag == nil {
		common.Warnf("New tag '%v'.", tagName)
		tag, err = db.AddTag(tagName)
		if err != nil {
			return nil, nil, err
		}
	}

	fileTag, err := db.FileTagByFileIdAndTagId(fileId, tag.Id)
	if err != nil {
		return nil, nil, err
	}

	if fileTag == nil {
		_, err := db.AddFileTag(fileId, tag.Id)
		if err != nil {
			return nil, nil, err
		}
	}

	return tag, fileTag, nil
}

func (command TagCommand) addFile(db *database.Database, path string) (*database.File, error) {
	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, err
	}

	file, err := db.FileByPath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	modTime := info.ModTime().UTC()

	if file == nil {
		// new file

		if !info.IsDir() {
			duplicateCount, err := db.FileCountByFingerprint(fingerprint)
			if err != nil {
				return nil, err
			}

			if duplicateCount > 0 {
				common.Warn("'" + common.RelPath(path) + "' is a duplicate of previously tagged files.")
			}
		}

		file, err = db.AddFile(path, fingerprint, modTime)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			fsFile, err := os.Open(file.Path())
			if err != nil {
				return nil, err
			}
			defer fsFile.Close()

			dirFilenames, err := fsFile.Readdirnames(0)
			if err != nil {
				return nil, err
			}

			for _, dirFilename := range dirFilenames {
				command.addFile(db, filepath.Join(path, dirFilename))
			}
		}
	} else {
		// existing file

		if file.ModTimestamp.Unix() != modTime.Unix() {
			db.UpdateFile(file.Id, file.Path(), fingerprint, modTime)
		}
	}

	return file, nil
}
