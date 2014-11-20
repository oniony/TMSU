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
	"path/filepath"
	"sort"
	"strings"
	"tmsu/common/log"
	"tmsu/common/path"
	"tmsu/entities"
	"tmsu/query"
	"tmsu/storage"
)

var FilesCommand = Command{
	Name:     "files",
	Aliases:  []string{"query"},
	Synopsis: "List files with particular tags",
	Usages:   []string{"tmsu files [OPTION]... [QUERY]"},
	Description: `Lists the files in the database that match the QUERY specified. If no query is specified, all files in the database are listed.

QUERY may contain tag names to match, operators and parentheses. Operators are: and or not == != < > <= >=.

Queries are run against the database so the results may not reflect the current state of the filesystem. Only tagged files are matched: to identify untagged files use the 'untagged' subcommand.

Note: Your shell may use some punctuation (e.g. < and >) for its own purposes.  Either enclose the query in quotation marks, escape the problematic characters or use the equivalent text operators: == eq, != ne, < lt, > gt, <= le, >= ge.`,
	Examples: []string{"$ tmsu files music mp3  # files with both 'music' and 'mp3'",
		"$ tmsu files music and mp3  # same query but with explicit 'and'",
		"$ tmsu files music and not mp3",
		`$ tmsu files \"music and (mp3 or flac)"}`,
		`$ tmsu files "year == 2014"  # tagged 'year' with a value '2014'`,
		`$ tmsu files "year < 2014" # tagged 'year' with values under '2014'`,
		`$ tmsu files year lt 2014  # same query but using textual operator`,
		`$ tmsu files year  # tagged 'year' (any or no value)`,
		`$ tmsu files --top music  # don't list individual files if directory is tagged`,
		`$ tmsu files --path=/home/bob music  # tagged 'music' under /home/bob`},
	Options: Options{{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--top", "-t", "list only the top-most matching items (exclude files under matching directories)", false, ""},
		{"--leaf", "-l", "list only the leaf items (files and directories without tagged contents)", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""},
		{"--path", "-p", "list only items under PATH", true, ""},
		{"--explicit", "-e", "list only explicitly tagged files", false, ""}},
	Exec: filesExec,
}

func filesExec(options Options, args []string) error {
	dirOnly := options.HasOption("--directory")
	fileOnly := options.HasOption("--file")
	topOnly := options.HasOption("--top")
	leafOnly := options.HasOption("--leaf")
	print0 := options.HasOption("--print0")
	showCount := options.HasOption("--count")
	hasPath := options.HasOption("--path")
	explicitOnly := options.HasOption("--explicit")

	absPath := ""
	if hasPath {
		relPath := options.Get("--path").Argument

		var err error
		absPath, err = filepath.Abs(relPath)
		if err != nil {
			fmt.Println("could not get absolute path of '%v': %v'", relPath, err)
		}
	}

	queryText := strings.Join(args, " ")
	return listFilesForQuery(queryText, absPath, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount, explicitOnly)
}

// unexported

func listFilesForQuery(queryText, path string, dirOnly, fileOnly, topOnly, leafOnly, print0, showCount, explicitOnly bool) error {
	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	log.Info(2, "parsing query")

	expression, err := query.Parse(queryText)
	if err != nil {
		return fmt.Errorf("could not parse query: %v", err)
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
		return errBlank
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

func containsTag(tags []string, tag string) bool {
	for _, iteratedTag := range tags {
		if iteratedTag == tag {
			return true
		}
	}

	return false
}
