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
	"strings"
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
  * The logical operators: 'and', 'or' and 'not'
  * The comparison operators: '=', '>', '<', '>=' and '<='
  * Parentheses: '(' and ')'

The 'and' operator may be omitted for brevity, e.g. 'chalk cheese' is
interpretted as 'chalk and cheese'.

The comparison operators are used to match on the values of tags. For example,
'country=uk' will only match files tagged 'country=uk' whilst 'year<=2014' will
match files tagged 'year=2014', 'year=2000', &c.

Note: Your shell may try to interpret some of the punctuation, e.g. most shells
will interpret the '<' and '>' operators as stream redirects. Enclosing the
query in quotation marks is often sufficient to avoid this but some characters
may need to be escaped (normally with a backslash).

Examples:

    $ tmsu files music mp3                # files with both 'music' and 'mp3'
    $ tmsu files music and mp3            # same query but with explicit 'and'
    $ tmsu files music and not mp3
    $ tmsu files "music and (mp3 or flac)"
    $ tmsu files year=2014                # tagged 'year' with a value '2014'
    $ tmsu files "year<2014"              # tagged 'year' with values under '2014'
    $ tmsu files year                     # tagged 'year' (any or no value)`,
	Options: Options{{"--all", "-a", "list the complete set of tagged files", false, ""},
		{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--top", "-t", "list only the top-most matching items (excludes the contents of matching directories)", false, ""},
		{"--leaf", "-l", "list only the leaf items (files and empty directories)", false, ""},
		{"--recursive", "-r", "read all files on the file-system under each matching directory, recursively", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""}},
	Exec: filesExec,
}

func filesExec(options Options, args []string) error {
	dirOnly := options.HasOption("--directory")
	fileOnly := options.HasOption("--file")
	topOnly := options.HasOption("--top")
	leafOnly := options.HasOption("--leaf")
	recursive := options.HasOption("--recursive")
	print0 := options.HasOption("--print0")
	showCount := options.HasOption("--count")

	if options.HasOption("--all") {
		return listAllFiles(dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount)
	}

	queryText := strings.Join(args, " ")
	return listFilesForQuery(queryText, dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount)
}

// unexported

func listAllFiles(dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount bool) error {
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

	return listFiles(files, dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount)
}

func listFilesForQuery(queryText string, dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount bool) error {
	if queryText == "" {
		return fmt.Errorf("query must be specified. Use --all to show all files.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

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

	log.Info(2, "checking value names")

	valueNames := query.ValueNames(expression)
	values, err := store.ValuesByNames(valueNames)
	for _, valueName := range valueNames {
		if !values.ContainsName(valueName) {
			log.Warnf("no such value '%v'.", valueName)
			wereErrors = true
			continue
		}
	}

	if wereErrors {
		return blankError
	}

	log.Info(2, "querying database")

	files, err := store.QueryFiles(expression)
	if err != nil {
		return fmt.Errorf("could not query files: %v", err)
	}

	if err = listFiles(files, dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount); err != nil {
		return err
	}

	return nil
}

func listFiles(files entities.Files, dirOnly, fileOnly, topOnly, leafOnly, recursive, print0, showCount bool) error {
	tree := path.NewTree()
	for _, file := range files {
		tree.Add(file.Path(), file.IsDir)
	}

	if topOnly {
		tree = tree.TopLevel()
	} else {
		if recursive {
			fsFiles, err := path.Enumerate(tree.TopLevel().Paths())
			if err != nil {
				return err
			}

			for _, fsFile := range fsFiles {
				tree.Add(fsFile.Path, fsFile.IsDir)
			}
		}
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

	if showCount {
		fmt.Println(len(tree.Paths()))
	} else {
		for _, absPath := range tree.Paths() {
			relPath := path.Rel(absPath)

			if print0 {
				fmt.Printf("%v\000", relPath)
			} else {
				fmt.Println(relPath)
			}
		}
	}

	return nil
}
