// Copyright 2011-2017 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package query

import (
	"fmt"
)

func Parse(query string) (Expression, error) {
	scanner := NewScanner(query)
	parser := NewParser(scanner)

	return parser.Parse()
}

// Creates an 'and' expression for all the tag names specified
func HasAll(tagNames []string) Expression {
	if len(tagNames) == 0 {
		return EmptyExpression{}
	}

	var expression Expression = TagExpression{tagNames[0]}

	for _, tagName := range tagNames[1:] {
		expression = AndExpression{expression, TagExpression{tagName}}
	}

	return expression
}

// Retrieves the set of tag names from an expression
func TagNames(expression Expression) ([]string, error) {
	names := make([]string, 0, 10)
	return tagNames(expression, names)
}

// Retrieves the set of value names from an expression where the name is matched on exactly
func ExactValueNames(expression Expression) ([]string, error) {
	names := make([]string, 0, 10)

	return exactValueNames(expression, names)
}

// unexported

func tagNames(expression Expression, names []string) ([]string, error) {
	var err error

	switch exp := expression.(type) {
	case EmptyExpression:
		// nowt
	case TagExpression:
		names = append(names, exp.Name)
	case NotExpression:
		names, err = tagNames(exp.Operand, names)
		if err != nil {
			return nil, err
		}
	case AndExpression:
		names, err = tagNames(exp.LeftOperand, names)
		if err != nil {
			return nil, err
		}

		names, err = tagNames(exp.RightOperand, names)
		if err != nil {
			return nil, err
		}
	case OrExpression:
		names, err = tagNames(exp.LeftOperand, names)
		if err != nil {
			return nil, err
		}

		names, err = tagNames(exp.RightOperand, names)
		if err != nil {
			return nil, err
		}
	case ComparisonExpression:
		names = append(names, exp.Tag.Name)
	default:
		return nil, fmt.Errorf("unsupported token type '%t'", exp)
	}

	return names, nil
}

func exactValueNames(expression Expression, names []string) ([]string, error) {
	var err error

	switch exp := expression.(type) {
	case EmptyExpression:
		// nowt
	case TagExpression:
		// nowt
	case NotExpression:
		names, err = exactValueNames(exp.Operand, names)
		if err != nil {
			return nil, err
		}
	case AndExpression:
		names, err = exactValueNames(exp.LeftOperand, names)
		if err != nil {
			return nil, err
		}

		names, err = exactValueNames(exp.RightOperand, names)
		if err != nil {
			return nil, err
		}
	case OrExpression:
		names, err = exactValueNames(exp.LeftOperand, names)
		if err != nil {
			return nil, err
		}

		names, err = exactValueNames(exp.RightOperand, names)
		if err != nil {
			return nil, err
		}
	case ComparisonExpression:
		switch exp.Operator {
		case "=", "==", "!=":
			names = append(names, exp.Value.Name)
		case "<", ">", "<=", ">=":
			// do nowt
		default:
			return nil, fmt.Errorf("unsupported operator '%v'", exp.Operator)
		}
	default:
		return nil, fmt.Errorf("unsupported token type '%t'", exp)
	}

	return names, nil
}
