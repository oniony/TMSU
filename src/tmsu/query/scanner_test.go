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

package query

import (
	"testing"
)

func TestSingleTag(test *testing.T) {
	scanner := NewScanner("cheese")

	lookAhead, err := scanner.LookAhead()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(lookAhead, "cheese", test)

	token, err := scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "cheese", test)

	lookAhead, err = scanner.LookAhead()
	if err != nil {
		test.Fatal(err)
	}
	validateEnd(lookAhead, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateEnd(token, test)
}

func TestSingleTagInParenthesese(test *testing.T) {
	scanner := NewScanner("(cheese)")

	token, err := scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateOpenParen(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "cheese", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateCloseParen(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateEnd(token, test)
}

func TestSingleTagWithValue(test *testing.T) {
	scanner := NewScanner("filling=cheese")

	token, err := scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "filling", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateEqualOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "cheese", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateEnd(token, test)
}

func TestComplexQuery(test *testing.T) {
	scanner := NewScanner("not cheese and (peas or sweetcorn) and not beans")

	token, err := scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateNotOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "cheese", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateAndOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateOpenParen(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "peas", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateOrOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "sweetcorn", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateCloseParen(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateAndOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateNotOperator(token, test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateSymbolToken(token, "beans", test)

	token, err = scanner.Next()
	if err != nil {
		test.Fatal(err)
	}
	validateEnd(token, test)
}

// unexported

func validateSymbolToken(token Token, expectedName string, test *testing.T) {
	tag := token.(SymbolToken)
	if tag.name != expectedName {
		test.Fatalf("Expected symbol '%v' but was '%v'.", expectedName, tag.name)
	}
}

func validateNotOperator(token Token, test *testing.T) {
	switch token.(type) {
	case NotOperatorToken:
		return
	default:
		test.Fatalf("Expected 'not' operator but was '%v'.", token)
	}
}

func validateAndOperator(token Token, test *testing.T) {
	switch token.(type) {
	case AndOperatorToken:
		return
	default:
		test.Fatalf("Expected 'and' operator but was '%v'.", token)
	}
}

func validateOrOperator(token Token, test *testing.T) {
	switch token.(type) {
	case OrOperatorToken:
		return
	default:
		test.Fatalf("Expected 'or' operator but was '%v'.", token)
	}
}

func validateEqualOperator(token Token, test *testing.T) {
	switch token.(type) {
	case EqualOperatorToken:
		return
	default:
		test.Fatalf("Expected '=' operator but was '%v'.", token)
	}
}

func validateOpenParen(token Token, test *testing.T) {
	switch token.(type) {
	case OpenParenToken:
		return
	default:
		test.Fatalf("Expected '(' but was '%v'.", token)
	}
}

func validateCloseParen(token Token, test *testing.T) {
	switch token.(type) {
	case CloseParenToken:
		return
	default:
		test.Fatalf("Expected ')' but was '%v'.", token)
	}
}

func validateEnd(token Token, test *testing.T) {
	switch token.(type) {
	case EndToken:
		return
	default:
		test.Fatalf("Expected end but was '%v'.", token)
	}
}
