/*
Copyright 2011-2013 Paul Ruane.

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

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("'%v': could not stat: %v", path, err)
	}

	return info.IsDir(), nil
}

func RelPath(path string) string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return path
	}

	if path == workingDirectory {
		return "."
	}

	if strings.HasPrefix(path, workingDirectory+string(filepath.Separator)) {
		return path[len(workingDirectory)+1:]
	}

	return path
}

func Join(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(dir, path)
}

func TopLevelPaths(paths []string) ([]string, error) {
	root := newNode("/", false)

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("'%v': could not get absolute path: %v", path, err)
		}

		pathParts := strings.Split(absPath, string(filepath.Separator))

		currentNode := root
		partCount := len(pathParts)
		for index, pathPart := range pathParts[1:] {
			top := index == (partCount - 2)
			node, found := currentNode.nodes[pathPart]
			if !found {
				node = newNode(pathPart, top)
				currentNode.nodes[pathPart] = node
			} else {
				if node.top {
					break
				} else {
					if top {
						node.top = true
						node.nodes = nil
					}
				}
			}

			currentNode = node
		}
	}

	return root.walk(), nil
}

// -

type node struct {
	name  string
	nodes map[string]*node
	top   bool
}

func newNode(name string, top bool) *node {
	return &node{name, make(map[string]*node, 0), top}
}

func (node *node) walk() []string {
	paths := node.walkImpl(make([]string, 0, 10), "")
	sort.Strings(paths)

	return paths
}

func (node *node) walkImpl(paths []string, path string) []string {
	path = filepath.Join(path, node.name)

	if node.top {
		return append(paths, path)
	}

	for _, childNode := range node.nodes {
		paths = childNode.walkImpl(paths, path)
	}

	return paths
}
