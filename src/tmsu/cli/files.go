// Copyright 2011-2015 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cli

import (
	"fmt"
	"path/filepath"
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

Note: Your shell may use some punctuation (e.g. < and >) for its own purposes. Either enclose the query in quotation marks, escape the problematic characters or use the equivalent text operators: == eq, != ne, < lt, > gt, <= le, >= ge.`,
	Examples: []string{"$ tmsu files music mp3  # files with both 'music' and 'mp3'",
		"$ tmsu files music and mp3  # same query but with explicit 'and'",
		"$ tmsu files music and not mp3",
		`$ tmsu files \"music and (mp3 or flac)"}`,
		`$ tmsu files "year == 2015"  # tagged 'year' with a value '2015'`,
		`$ tmsu files "year < 2015" # tagged 'year' with values under '2015'`,
		`$ tmsu files year lt 2015  # same query but using textual operator`,
		`$ tmsu files year  # tagged 'year' (any or no value)`,
		`$ tmsu files --path=/home/bob music  # tagged 'music' under /home/bob`},
	Options: Options{{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""},
		{"--path", "-p", "list only items under PATH", true, ""},
		{"--explicit", "-e", "list only explicitly tagged files", false, ""},
		{"--sort", "-s", "sort output: id, none, name, size, time", true, ""}},
	Exec: filesExec,
}

func filesExec(store *storage.Storage, options Options, args []string) error {
	dirOnly := options.HasOption("--directory")
	fileOnly := options.HasOption("--file")
	print0 := options.HasOption("--print0")
	showCount := options.HasOption("--count")
	hasPath := options.HasOption("--path")
	explicitOnly := options.HasOption("--explicit")

	sort := "name"
	if options.HasOption("--sort") {
		sort = options.Get("--sort").Argument
	}

	absPath := ""
	if hasPath {
		relPath := options.Get("--path").Argument

		var err error
		absPath, err = filepath.Abs(relPath)
		if err != nil {
			fmt.Println("could not get absolute path of '%v': %v'", relPath, err)
		}
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	queryText := strings.Join(args, " ")
	return listFilesForQuery(store, tx, queryText, absPath, dirOnly, fileOnly, print0, showCount, explicitOnly, sort)
}

// unexported

func listFilesForQuery(store *storage.Storage, tx *storage.Tx, queryText, path string, dirOnly, fileOnly, print0, showCount, explicitOnly bool, sort string) error {
	log.Info(2, "parsing query")

	expression, err := query.Parse(queryText)
	if err != nil {
		return fmt.Errorf("could not parse query: %v", err)
	}

	log.Info(2, "checking tag names")

	wereErrors := false

	tagNames := query.TagNames(expression)
	tags, err := store.TagsByNames(tx, tagNames)
	for _, tagName := range tagNames {
		if !tags.ContainsName(tagName) {
			log.Warnf("no such tag '%v'", tagName)
			wereErrors = true
			continue
		}
	}

	//TODO only check value names for == or != comparisons
	/*
		valueNames := query.ValueNames(expression)
		values, err := store.ValuesByNames(tx, valueNames)
		for _, valueName := range valueNames {
			if !values.ContainsName(valueName) {
				log.Warnf("no such value '%v'", valueName)
				wereErrors = true
				continue
			}
		}
	*/

	if wereErrors {
		return errBlank
	}

	log.Info(2, "querying database")

	files, err := store.FilesForQuery(tx, expression, path, explicitOnly, sort)
	if err != nil {
		if strings.Index(err.Error(), "parser stack overflow") > -1 {
			return fmt.Errorf("the query is too complex (see the troubleshooting wiki for how to increase the stack size)")
		} else {
			return fmt.Errorf("could not query files: %v", err)
		}
	}

	if err = listFiles(tx, files, dirOnly, fileOnly, print0, showCount); err != nil {
		return err
	}

	return nil
}

func listFiles(tx *storage.Tx, files entities.Files, dirOnly, fileOnly, print0, showCount bool) error {
	relPaths := make([]string, 0, len(files))
	for _, file := range files {
		if fileOnly && file.IsDir {
			continue
		}
		if dirOnly && !file.IsDir {
			continue
		}

		absPath := file.Path()
		relPath := path.Rel(absPath)

		relPaths = append(relPaths, relPath)
	}

	if showCount {
		fmt.Println(len(relPaths))
	} else {
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
