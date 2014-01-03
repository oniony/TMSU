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

package path

import (
	"path/filepath"
	"sort"
	"strings"
)

type Tree struct {
	root *node
}

func NewTree() *Tree {
	return &Tree{newNode("/", false, true)}
}

// Adds a path to the tree
func (tree *Tree) Add(path string, isDir bool) {
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
			node = newNode(pathPart, isReal, true)
			currentNode.nodes[pathPart] = node
		} else {
			if isReal && !node.isReal {
				node.isReal = true
			}
		}

		currentNode.isDir = true
		currentNode = node
	}

	currentNode.isDir = isDir
}

// The set of paths in the tree
func (tree *Tree) Paths() []string {
	paths := make([]string, 0, 100)
	paths = tree.root.paths(paths, "")
	sort.Strings(paths)

	return paths
}

// Builds a top-level tree
func (tree *Tree) TopLevel() *Tree {
	resultTree := NewTree()
	tree.root.findTopLevel(resultTree.root)

	return resultTree
}

// Finds the leaf nodes added to the tree
func (tree *Tree) Leaves() *Tree {
	resultTree := NewTree()
	tree.root.findLeaves(resultTree.root)

	return resultTree
}

// Finds the nodes that are files
func (tree *Tree) Files() *Tree {
	resultTree := NewTree()
	tree.root.findFiles(resultTree.root)

	return resultTree
}

// Finds the nodes that are directories
func (tree *Tree) Directories() *Tree {
	resultTree := NewTree()
	tree.root.findDirectories(resultTree.root)

	return resultTree
}

type node struct {
	name   string
	nodes  map[string]*node
	isReal bool
	isDir  bool
}

func newNode(name string, isReal bool, isDir bool) *node {
	return &node{name, make(map[string]*node, 0), isReal, isDir}
}

func (node *node) paths(paths []string, prefix string) []string {
	if node.isReal {
		paths = append(paths, filepath.Join(prefix, node.name))
	}

	for _, childNode := range node.nodes {
		paths = childNode.paths(paths, filepath.Join(prefix, node.name))
	}

	return paths
}

func (node *node) findTopLevel(resultNode *node) {
	resultNode.isReal = node.isReal
	if node.isReal {
		return
	}

	for _, childNode := range node.nodes {
		resultChildNode := newNode(childNode.name, false, childNode.isDir)
		resultNode.nodes[childNode.name] = resultChildNode

		childNode.findTopLevel(resultChildNode)
	}
}

func (node *node) findLeaves(resultNode *node) {
	resultNode.isReal = node.isReal && len(node.nodes) == 0

	for _, childNode := range node.nodes {
		resultChildNode := newNode(childNode.name, false, childNode.isDir)
		resultNode.nodes[childNode.name] = resultChildNode

		childNode.findLeaves(resultChildNode)
	}
}

func (node *node) findFiles(resultNode *node) {
	resultNode.isReal = node.isReal && !node.isDir

	for _, childNode := range node.nodes {
		resultChildNode := newNode(childNode.name, false, childNode.isDir)
		resultNode.nodes[childNode.name] = resultChildNode

		childNode.findFiles(resultChildNode)
	}
}

func (node *node) findDirectories(resultNode *node) {
	resultNode.isReal = node.isReal && node.isDir

	for _, childNode := range node.nodes {
		resultChildNode := newNode(childNode.name, false, childNode.isDir)
		resultNode.nodes[childNode.name] = resultChildNode

		childNode.findDirectories(resultChildNode)
	}
}
