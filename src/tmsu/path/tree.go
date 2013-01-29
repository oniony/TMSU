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

type Node struct {
	Name  string
	Nodes map[string]*Node
	Root  bool
}

// Create a new tree
func NewTree() *Node {
	return NewNode("", false)
}

// Create a new tree node
func NewNode(name string, root bool) *Node {
	return &Node{name, make(map[string]*Node, 0), root}
}

func (node *Node) Add(path string) {
	pathParts := strings.Split(path, string(filepath.Separator))

	currentNode := node
	partCount := len(pathParts)
	for index, pathPart := range pathParts {
		if pathPart == "" {
			pathPart = "/"
		}

		root := index == (partCount - 1)
		node, found := currentNode.Nodes[pathPart]
		if !found {
			node = NewNode(pathPart, root)
			currentNode.Nodes[pathPart] = node
		} else {
			if node.Root {
				break
			} else {
				if root {
					node.Root = true
					node.Nodes = nil
				}
			}
		}

		currentNode = node
	}
}

func (node *Node) Roots() []string {
	roots := node.findRoots(make([]string, 0, 10), "")

	sort.Strings(roots)

	return roots
}

func (node *Node) findRoots(paths []string, path string) []string {
	path = filepath.Join(path, node.Name)

	if node.Root {
		return append(paths, path)
	}

	for _, childNode := range node.Nodes {
		paths = childNode.findRoots(paths, path)
	}

	return paths
}
