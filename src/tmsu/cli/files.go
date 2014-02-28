/*
Copyright 2011-2014 Paul Ruane.

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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tmsu/common/log"
	"tmsu/common/path"
	"tmsu/entities"
	"tmsu/query"
	"tmsu/storage"
)

var FilesCommand = Command{
	Name:     "files",
	Synopsis: "List files with particular tags",
	Description: `tmsu files [OPTION]... QUERY 

Lists the files that match the QUERY specified.

QUERY may contain:

  * Tag names to match
  * The logical operators: and or not
  * The comparison operators: =   !=   >   <   >=   <=
                              eq  ne   gt  lt  ge   le
  * Parentheses: ( )

The 'and' operator may be omitted for brevity, e.g. 'chalk cheese' is
interpretted as 'chalk and cheese'.

Note: Your shell may interpret some of the punctuation, e.g. most shells will
interpret the '<' and '>' operators as stream redirects. Either enclose the
query in quotation marks, escape the problematic characters or use the text
versions, e.g. 'eq' for '='.

Examples:

    $ tmsu files music mp3                # files with both 'music' and 'mp3'
    $ tmsu files music and mp3            # same query but with explicit 'and'
    $ tmsu files music and not mp3
    $ tmsu files "music and (mp3 or flac)"

    $ tmsu files year = 2014              # tagged 'year' with a value '2014'
    $ tmsu files "year < 2014"            # tagged 'year' with values under '2014'
    $ tmsu files year                     # tagged 'year' (any or no value)

    $ tmsu files --top music              # don't list individual files if directory is tagged
    $ tmsu files --path=/home/bob music   # tagged 'music' under /home/bob`,
	Options: Options{{"--all", "-a", "list the complete set of tagged files", false, ""},
		{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--top", "-t", "list only the top-most matching items (exclude files under matching directories)", false, ""},
		{"--leaf", "-l", "list only the leaf items (files and directories without tagged contents)", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""},
		{"--path", "-p", "list only items under PATH", true, ""},
		{"--untagged", "-u", "combined with --path, lists untagged files under PATH", false, ""},
		{"--explicit", "-e", "list only explicitly tagged files", false, ""}},
	Exec: filesExec,
}

func filesExec(options Options, args []string) error {
	absPath := ""
	if options.HasOption("--path") {
		pth := options.Get("--path").Argument

		var err error
		absPath, err = filepath.Abs(pth)
		if err != nil {
			fmt.Println("could not get absolute path of '%v': %v'", pth, err)
		}
	}

	dirOnly := options.HasOption("--directory")
	fileOnly := options.HasOption("--file")
	topOnly := options.HasOption("--top")
	leafOnly := options.HasOption("--leaf")
	print0 := options.HasOption("--print0")
	showCount := options.HasOption("--count")
	untagged := options.HasOption("--untagged")
	explicitOnly := options.HasOption("--explicit")

	if options.HasOption("--all") {
		return listAllFiles(dirOnly, fileOnly, topOnly, leafOnly, print0, showCount)
	}

	queryText := strings.Join(args, " ")
	return listFilesForQuery(queryText, absPath, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount, untagged, explicitOnly)
}

// unexported

func listAllFiles(dirOnly, fileOnly, topOnly, leafOnly, print0, showCount bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "retrieving all files from database.")

	files, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve files: %v", err)
	}

	return listFiles(files, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount)
}

func listFilesForQuery(queryText, path string, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount, untagged, explicitOnly bool) error {
	if queryText == "" {
		return fmt.Errorf("query must be specified. Use --all to show all files.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if untagged && path != "" {
		log.Info(2, "temporarily adding untagged files")

		if err := store.Begin(); err != nil {
			return fmt.Errorf("could not begin transaction: %v", err)
		}
		defer store.Rollback()

		if err := addUntaggedFiles(store, path); err != nil {
			return err
		}
	}

	log.Info(2, "parsing query")

	expression, err := query.Parse(queryText)
	if err != nil {
		return err
	}

	log.Info(2, "checking tag names")

	wereErrors := false

	tagNames := query.TagNames(expression)
	tags, err := store.TagsByNames(tagNames)
	for _, tagName := range tagNames {
		if !tags.ContainsName(tagName) {
			log.Warnf("no such tag '%v'.", tagName)
			wereErrors = true
			continue
		}
	}

	if wereErrors {
		return blankError
	}

	log.Info(2, "querying database")

	files, err := store.QueryFiles(expression, path, explicitOnly)
	if err != nil {
		return fmt.Errorf("could not query files: %v", err)
	}

	if err = listFiles(files, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount); err != nil {
		return err
	}

	return nil
}

func listFiles(files entities.Files, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount bool) error {
	tree := path.NewTree()
	for _, file := range files {
		tree.Add(file.Path(), file.IsDir)
	}

	if topOnly {
		tree = tree.TopLevel()
	}

	if leafOnly {
		tree = tree.Leaves()
	}

	if fileOnly {
		tree = tree.Files()
	}

	if dirOnly {
		tree = tree.Directories()
	}

	absPaths := tree.Paths()

	if showCount {
		fmt.Println(len(absPaths))
	} else {
		relPaths := make([]string, len(absPaths))
		for index, absPath := range absPaths {
			relPaths[index] = path.Rel(absPath)
		}
		sort.Strings(relPaths)

		for _, relPath := range relPaths {
			if print0 {
				fmt.Printf("%v\000", relPath)
			} else {
				fmt.Println(relPath)
			}
		}
	}

	return nil
}

func addUntaggedFiles(store *storage.Storage, path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%v: could not stat: %v", path, err)
	}

	if !stat.IsDir() {
		return nil
	}

	return addUntaggedFilesRecursive(store, path, stat)
}

func addUntaggedFilesRecursive(store *storage.Storage, path string, stat os.FileInfo) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%v: could not open directory: %v", path, err)
	}

	stats, err := file.Readdir(0)
	if err != nil {
		return fmt.Errorf("%v: could not enumerate directory: %v", path, err)
	}

	file.Close()

	for _, stat := range stats {
		entryPath := path + string(filepath.Separator) + stat.Name()
		_, _ = store.AddFile(entryPath, "", time.Time{}, 0, false)

		if !stat.IsDir() {
			continue
		}

		if err := addUntaggedFilesRecursive(store, entryPath, stat); err != nil {
			return err
		}
	}

	return nil
}

func containsTag(tags []string, tag string) bool {
	for _, iteratedTag := range tags {
		if iteratedTag == tag {
			return true
		}
	}

	return false
}
