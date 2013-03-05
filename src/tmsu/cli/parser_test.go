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

package cli

import (
	"testing"
)

func TestParseVanillaArguments(test *testing.T) {
	parser := NewParser(Options{}, make(map[CommandName]Command))

	commandName, options, arguments, err := parser.Parse([]string{"a", "b", "c"})
	if err != nil {
		test.Fatal(err)
	}
	if commandName != "a" {
		test.Fatalf("Expected command name of 'a' but was '%v'.", commandName)
	}
	if len(options) != 0 {
		test.Fatalf("Expected zero options but were %v.", len(options))
	}
	if len(arguments) != 2 {
		test.Fatalf("Expected two arguments but were %v.", len(arguments))
	}
	if arguments[0] != "b" || arguments[1] != "c" {
		test.Fatalf("Expected arguments of 'b' and 'c' but were '%v' and '%v'.", arguments[0], arguments[1])
	}
}

func TestParseGlobalOptions(test *testing.T) {
	parser := NewParser(Options{Option{"--verbose", "-v", "verbose", false, ""}}, make(map[CommandName]Command))

	commandName, options, arguments, err := parser.Parse([]string{"--verbose", "a", "b"})
	if err != nil {
		test.Fatal(err)
	}
	if len(options) != 1 {
		test.Fatalf("Expected one global option but were %v.", len(globalOptions))
	}
	if options[0].LongName != "--verbose" {
		test.Fatalf("Expected option long name of '--verbose' but was '%v'.", globalOptions[0])
	}
	if options[0].ShortName != "-v" {
		test.Fatalf("Expected option short name of '-v' but was '%v'.", globalOptions[0])
	}
	if commandName != "a" {
		test.Fatalf("Expected command name of 'a' but was '%v'.", commandName)
	}
	if len(arguments) != 1 {
		test.Fatalf("Expected one argument but were %v.", len(arguments))
	}
	if arguments[0] != "b" {
		test.Fatalf("Expected argument of 'b' but was '%v'", arguments[0])
	}
}

func TestInvalidGlobalOption(test *testing.T) {
	parser := NewParser(Options{}, make(map[CommandName]Command))

	_, _, _, err := parser.Parse([]string{"--invalid", "a", "b"})

	if err == nil {
		test.Fatal("Invalid option not identified.")
	}
}
