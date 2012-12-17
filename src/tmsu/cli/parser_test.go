/*
Copyright 2011-2012 Paul Ruane.

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
	parser := Create()

	globalOptions, commandName, commandOptions, args, err := parser.Parse([]string{"a", "b", "c"})
	if err != nil {
		test.Fatal(err)
	}
	if len(globalOptions) != 0 {
		test.Fatalf("Expected zero global options but were %v.", len(globalOptions))
	}
	if commandName != "a" {
		test.Fatalf("Expected command name of 'a' but was '%v'.", commandName)
	}
	if len(commandOptions) != 0 {
		test.Fatalf("Expected zero command options but were %v.", len(commandOptions))
	}
	if len(args) != 2 {
		test.Fatalf("Expected two arguments but were %v.", len(args))
	}
	if args[0] != "b" || args[1] != "c" {
		test.Fatalf("Expected arguments of 'b' and 'c' but were '%v' and '%v'.", args[0], args[1])
	}
}

func TestParseGlobalOptions(test *testing.T) {
	parser := Create()

	globalOptions, commandName, commandOptions, args, err := parser.Parse([]string{"--verbose", "a", "b"})
	if err != nil {
		test.Fatal(err)
	}
	if len(globalOptions) != 1 {
		test.Fatalf("Expected one global option but were %v.", len(globalOptions))
	}
	if globalOptions[0].LongName != "--verbose" {
		test.Fatalf("Expected global option long name of '--verbose' but was '%v'.", globalOptions[0])
	}
	if globalOptions[0].ShortName != "-v" {
		test.Fatalf("Expected global option short name of '-v' but was '%v'.", globalOptions[0])
	}
	if commandName != "a" {
		test.Fatalf("Expected command name of 'a' but was '%v'.", commandName)
	}
	if len(commandOptions) != 0 {
		test.Fatalf("Expected zero command options but were %v.", len(commandOptions))
	}
	if len(args) != 1 {
		test.Fatalf("Expected one arguments but were %v.", len(args))
	}
	if args[0] != "b" {
		test.Fatalf("Expected argument of 'b' but was '%v'", args[0])
	}
}

func TestInvalidGlobalOption(test *testing.T) {
	parser := Create()

	_, _, _, _, err := parser.Parse([]string{"--invalid", "a", "b"})

	if err == nil {
		test.Fatal("Invalid option not identified.")
	}
}
