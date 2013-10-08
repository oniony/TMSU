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
	"fmt"
	"testing"
)

func TestTagParsing(test *testing.T) {
	scanner := NewScanner("cheese")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	_ = expression.(TagExpression)
}

func TestNotParsing(test *testing.T) {
	scanner := NewScanner("not cheese")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	not := expression.(NotExpression)
	_ = not.Operand.(TagExpression)
}

func TestImplicitAndParsing(test *testing.T) {
	scanner := NewScanner("cheese tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := expression.(AndExpression)
	_ = and.LeftOperand.(TagExpression)
	_ = and.RightOperand.(TagExpression)
}

func TestImplicitNotAndParsing(test *testing.T) {
	scanner := NewScanner("not cheese tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := expression.(AndExpression)
	not := and.LeftOperand.(NotExpression)
	_ = not.Operand.(TagExpression)
	_ = and.RightOperand.(TagExpression)
}

func TestImplicitAndNotParsing(test *testing.T) {
	scanner := NewScanner("cheese not tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := expression.(AndExpression)
	_ = and.LeftOperand.(TagExpression)
	not := and.RightOperand.(NotExpression)
	_ = not.Operand.(TagExpression)
}

func TestAndParsing(test *testing.T) {
	scanner := NewScanner("cheese and tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := expression.(AndExpression)
	_ = and.LeftOperand.(TagExpression)
	_ = and.RightOperand.(TagExpression)
}

func TestOrParsing(test *testing.T) {
	scanner := NewScanner("cheese or tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or := expression.(OrExpression)
	_ = or.LeftOperand.(TagExpression)
	_ = or.RightOperand.(TagExpression)
}

func TestOrOrParsing(test *testing.T) {
	scanner := NewScanner("cheese or tomato or sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or1 := expression.(OrExpression)
	or2 := or1.LeftOperand.(OrExpression)
	_ = or1.RightOperand.(TagExpression)
	_ = or2.LeftOperand.(TagExpression)
	_ = or2.RightOperand.(TagExpression)
}

func dump(expression Expression) {
	dumpBranch(expression)
	fmt.Println()
}

func dumpBranch(expression Expression) {
	switch exp := expression.(type) {
	case TagExpression:
		fmt.Printf(exp.Name)
	case NotExpression:
		fmt.Printf("Not(")
		dumpBranch(exp.Operand)
		fmt.Printf(")")
	case AndExpression:
		fmt.Printf("And(")
		dumpBranch(exp.LeftOperand)
		fmt.Printf(", ")
		dumpBranch(exp.RightOperand)
		fmt.Printf(")")
	case OrExpression:
		fmt.Printf("Or(")
		dumpBranch(exp.LeftOperand)
		fmt.Printf(", ")
		dumpBranch(exp.RightOperand)
		fmt.Printf(")")
	}
}
