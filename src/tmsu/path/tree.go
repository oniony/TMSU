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

package path

import (
	"path/filepath"
	"sort"
	"strings"
)

// Finds the non-nested nodes added to the tree
func NonNested(paths []string) ([]string, error) {
	tree := buildTree(paths)

	nonNested := make([]string, 0, 100)
	nonNested = tree.root.findNonNested(nonNested, "")
	sort.Strings(nonNested)

	return nonNested, nil
}

// Finds the leaf nodes added to the tree
func Leaves(paths []string) ([]string, error) {
	tree := buildTree(paths)

	leaves := make([]string, 0, 100)
	leaves = tree.root.findLeaves(leaves, "")
	sort.Strings(leaves)

	return leaves, nil
}

// -

type tree struct {
	root *node
}

type node struct {
	name   string
	nodes  map[string]*node
	isReal bool
}

func buildTree(paths []string) *tree {
	tree := tree{newNode("/", false)}

	for _, path := range paths {
		tree.add(path)
	}

	return &tree
}

func newNode(name string, isReal bool) *node {
	return &node{name, make(map[string]*node, 0), isReal}
}

func (tree *tree) add(path string) {
	pathParts := strings.Split(path, string(filepath.Separator))

	currentNode := tree.root
	partCount := len(pathParts)
	for index, pathPart := range pathParts {
		isReal := index == partCount-1

		if pathPart == "" {
			pathPart = "/"
		}

		node, found := currentNode.nodes[pathPart]
		if !found {
			node = newNode(pathPart, isReal)
			currentNode.nodes[pathPart] = node
		} else {
			if isReal && !node.isReal {
				node.isReal = true
			}
		}

		currentNode = node
	}
}

func (node *node) findNonNested(paths []string, path string) []string {
	path = filepath.Join(path, node.name)

	if node.isReal {
		return append(paths, path)
	}

	for _, childNode := range node.nodes {
		paths = childNode.findNonNested(paths, path)
	}

	return paths
}

func (node *node) findLeaves(paths []string, path string) []string {
	path = filepath.Join(path, node.name)

	if len(node.nodes) == 0 {
		return append(paths, path)
	}

	for _, childNode := range node.nodes {
		paths = childNode.findLeaves(paths, path)
	}

	return paths
}
