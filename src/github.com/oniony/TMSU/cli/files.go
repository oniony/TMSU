// Copyright 2011-2017 Paul Ruane.

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
	"github.com/oniony/TMSU/common/log"
	"github.com/oniony/TMSU/common/path"
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/query"
	"github.com/oniony/TMSU/storage"
	"path/filepath"
	"strings"
)

var FilesCommand = Command{
	Name:     "files",
	Aliases:  []string{"query"},
	Synopsis: "List files with particular tags",
	Usages:   []string{"tmsu files [OPTION]... [QUERY]"},
	Description: `Lists the files in the database that match the QUERY specified. If no query is specified, all files in the database are listed.

QUERY may contain tag names to match, operators and parentheses. Operators are: and or not == != < > <= >= eq ne lt gt le ge.

Queries are run against the database so the results may not reflect the current state of the filesystem. Only tagged files are matched: to identify untagged files use the 'untagged' subcommand.

Note: If your tag or value name contains whitespace, operators (e.g. '<') or parentheses ('(' or ')'), these must be escaped with a backslash '\', e.g. '\<tag\>' matches the tag name '<tag>'. Your shell, however, may use some punctuation for its own purposes: this can normally be avoided by enclosing the query in single quotation marks or by escaping the problem characters with a backslash.`,
	Examples: []string{"$ tmsu files music mp3  # files with both 'music' and 'mp3'",
		"$ tmsu files music and mp3  # same query but with explicit 'and'",
		"$ tmsu files music and not mp3",
		`$ tmsu files "music and (mp3 or flac)"`,
		`$ tmsu files "year == 2017"`,
		`$ tmsu files "year < 2017"`,
		`$ tmsu files year lt 2017`,
		`$ tmsu files year`,
		`$ tmsu files --path=/home/bob music`,
		`$ tmsu files 'contains\=equals'`,
		`$ tmsu files '\<tag\>'`},
	Options: Options{{"--directory", "-d", "list only items that are directories", false, ""},
		{"--file", "-f", "list only items that are files", false, ""},
		{"--print0", "-0", "delimit files with a NUL character rather than newline.", false, ""},
		{"--count", "-c", "lists the number of files rather than their names", false, ""},
		{"--path", "-p", "list only items under PATH", true, ""},
		{"--explicit", "-e", "list only explicitly tagged files", false, ""},
		{"--sort", "-s", "sort output: id, none, name, size, time", true, ""},
		{"--ignore-case", "-i", "ignore the case of tag and value names", false, ""}},
	Exec: filesExec,
}

// unexported

func filesExec(options Options, args []string, databasePath string) (error, warnings) {
	dirOnly := options.HasOption("--directory")
	fileOnly := options.HasOption("--file")
	print0 := options.HasOption("--print0")
	showCount := options.HasOption("--count")
	hasPath := options.HasOption("--path")
	explicitOnly := options.HasOption("--explicit")
	ignoreCase := options.HasOption("--ignore-case")

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
			return fmt.Errorf("could not get absolute path of '%v': %v'", relPath, err), nil
		}
	}

	store, err := openDatabase(databasePath)
	if err != nil {
		return err, nil
	}
	defer store.Close()

	tx, err := store.Begin()
	if err != nil {
		return err, nil
	}
	defer tx.Commit()

	queryText := strings.Join(args, " ")
	return listFilesForQuery(store, tx, queryText, absPath, dirOnly, fileOnly, print0, showCount, explicitOnly, ignoreCase, sort)
}

// unexported

func listFilesForQuery(store *storage.Storage, tx *storage.Tx, queryText, path string, dirOnly, fileOnly, print0, showCount, explicitOnly, ignoreCase bool, sort string) (error, warnings) {
	log.Info(2, "parsing query")

	expression, err := query.Parse(queryText)
	if err != nil {
		return fmt.Errorf("could not parse query: %v", err), nil
	}

	log.Info(2, "checking tag names")

	warnings := make(warnings, 0, 10)

	tagNames, err := query.TagNames(expression)
	if err != nil {
		return fmt.Errorf("could not identify tag names: %v", err), nil
	}

	tags, err := store.TagsByCasedNames(tx, tagNames, ignoreCase)
	for _, tagName := range tagNames {
		if err := entities.ValidateTagName(tagName); err != nil {
			warnings = append(warnings, err.Error())
			continue
		}

		if !tags.ContainsCasedName(tagName, ignoreCase) {
			warnings = append(warnings, fmt.Sprintf("no such tag '%v'", tagName))
			continue
		}
	}

	valueNames, err := query.ExactValueNames(expression)
	if err != nil {
		return fmt.Errorf("could not identify value names: %v", err), nil
	}

	values, err := store.ValuesByCasedNames(tx, valueNames, ignoreCase)
	for _, valueName := range valueNames {
		if err := entities.ValidateValueName(valueName); err != nil {
			warnings = append(warnings, err.Error())
			continue
		}

		if !values.ContainsCasedName(valueName, ignoreCase) {
			warnings = append(warnings, fmt.Sprintf("no such value '%v'", valueName))
			continue
		}
	}

	log.Info(2, "querying database")

	files, err := store.FilesForQuery(tx, expression, path, explicitOnly, ignoreCase, sort)
	if err != nil {
		if strings.Index(err.Error(), "parser stack overflow") > -1 {
			return fmt.Errorf("the query is too complex (see the troubleshooting wiki for how to increase the stack size)"), warnings
		}

		return fmt.Errorf("could not query files: %v", err), warnings
	}

	if err = listFiles(tx, files, dirOnly, fileOnly, print0, showCount); err != nil {
		return err, warnings
	}

	return nil, warnings
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
