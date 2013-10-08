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

	validateTag(expression, "cheese", test)
}

func TestNotParsing(test *testing.T) {
	scanner := NewScanner("not cheese")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	not := validateNot(expression)
	validateTag(not.Operand, "cheese", test)
}

func TestImplicitAndParsing(test *testing.T) {
	scanner := NewScanner("cheese tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := validateAnd(expression)
	validateTag(and.LeftOperand, "cheese", test)
	validateTag(and.RightOperand, "tomato", test)
}

func TestImplicitNotAndParsing(test *testing.T) {
	scanner := NewScanner("not cheese tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := validateAnd(expression)
	not := validateNot(and.LeftOperand)
	validateTag(not.Operand, "cheese", test)
	validateTag(and.RightOperand, "tomato", test)
}

func TestImplicitAndNotParsing(test *testing.T) {
	scanner := NewScanner("cheese not tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := validateAnd(expression)
	validateTag(and.LeftOperand, "cheese", test)
	not := validateNot(and.RightOperand)
	validateTag(not.Operand, "tomato", test)
}

func TestAndParsing(test *testing.T) {
	scanner := NewScanner("cheese and tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := validateAnd(expression)
	validateTag(and.LeftOperand, "cheese", test)
	validateTag(and.RightOperand, "tomato", test)
}

func TestOrParsing(test *testing.T) {
	scanner := NewScanner("cheese or tomato")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or := validateOr(expression)
	validateTag(or.LeftOperand, "cheese", test)
	validateTag(or.RightOperand, "tomato", test)
}

func TestOrOrParsing(test *testing.T) {
	scanner := NewScanner("cheese or tomato or sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or1 := validateOr(expression)
	or2 := validateOr(or1.LeftOperand)
	validateTag(or2.LeftOperand, "cheese", test)
	validateTag(or2.RightOperand, "tomato", test)
	validateTag(or1.RightOperand, "sweetcorn", test)
}

func TestAndAndParsing(test *testing.T) {
	scanner := NewScanner("cheese and tomato and sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and1 := validateAnd(expression)
	and2 := validateAnd(and1.LeftOperand)
	validateTag(and2.LeftOperand, "cheese", test)
	validateTag(and2.RightOperand, "tomato", test)
	validateTag(and1.RightOperand, "sweetcorn", test)
}

func TestAndOrParsing(test *testing.T) {
	scanner := NewScanner("cheese and tomato or sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or := validateOr(expression)
	and := validateAnd(or.LeftOperand)
	validateTag(and.LeftOperand, "cheese", test)
	validateTag(and.RightOperand, "tomato", test)
	validateTag(or.RightOperand, "sweetcorn", test)
}

func TestOrAndParsing(test *testing.T) {
	scanner := NewScanner("cheese or tomato and sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or := validateOr(expression)
	validateTag(or.LeftOperand, "cheese", test)
	and := validateAnd(or.RightOperand)
	validateTag(and.LeftOperand, "tomato", test)
	validateTag(and.RightOperand, "sweetcorn", test)
}

func TestParen(test *testing.T) {
	scanner := NewScanner("(cheese)")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	validateTag(expression, "cheese", test)
}

func TestDoubleParen(test *testing.T) {
	scanner := NewScanner("((cheese))")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	validateTag(expression, "cheese", test)
}

func TestNotParen(test *testing.T) {
	scanner := NewScanner("not (cheese)")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	not := validateNot(expression)
	validateTag(not.Operand, "cheese", test)
}

func TestParenNot(test *testing.T) {
	scanner := NewScanner("(not cheese)")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	not := validateNot(expression)
	validateTag(not.Operand, "cheese", test)
}

func TestParenAndOrParsing(test *testing.T) {
	scanner := NewScanner("(cheese and tomato) or sweetcorn")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	or := validateOr(expression)
	and := validateAnd(or.LeftOperand)
	validateTag(and.LeftOperand, "cheese", test)
	validateTag(and.RightOperand, "tomato", test)
	validateTag(or.RightOperand, "sweetcorn", test)
}

func TestAndParenOrParsing(test *testing.T) {
	scanner := NewScanner("cheese and (tomato or sweetcorn)")
	parser := NewParser(scanner)

	expression, err := parser.Parse()
	if err != nil {
		test.Fatal(err)
	}

	dump(expression)

	and := validateAnd(expression)
	validateTag(and.LeftOperand, "cheese", test)
	or := validateOr(and.RightOperand)
	validateTag(or.LeftOperand, "tomato", test)
	validateTag(or.RightOperand, "sweetcorn", test)
}

// unexported

func validateNot(expression Expression) NotExpression {
	return expression.(NotExpression)
}

func validateOr(expression Expression) OrExpression {
	return expression.(OrExpression)
}

func validateAnd(expression Expression) AndExpression {
	return expression.(AndExpression)
}

func validateTag(expression Expression, expectedName string, test *testing.T) TagExpression {
	tag := expression.(TagExpression)
	if tag.Name != expectedName {
		test.Fatalf("Expected '%v' tag but was '%v'.", tag.Name)
	}

	return tag
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
