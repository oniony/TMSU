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

type Command struct {
	Name        string
	Synopsis    string
	Description string
	Options     Options
	Exec        func(options Options, args []string) error
}

var commands = map[string]*Command{
	"copy":     &CopyCommand,
	"delete":   &DeleteCommand,
	"dupes":    &DupesCommand,
	"files":    &FilesCommand,
	"help":     &HelpCommand,
	"imply":    &ImplyCommand,
	"merge":    &MergeCommand,
	"relocate": &RelocateCommand,
	"rename":   &RenameCommand,
	"repair":   &RepairCommand,
	"stats":    &StatsCommand,
	"status":   &StatusCommand,
	"tag":      &TagCommand,
	"tags":     &TagsCommand,
	"untag":    &UntagCommand,
	"untagged": &UntaggedCommand,
	"values":   &ValuesCommand,
	"version":  &VersionCommand}
