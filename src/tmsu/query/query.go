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

package query

func Parse(query string) (Expression, error) {
	scanner := NewScanner(query)
	parser := NewParser(scanner)

	return parser.Parse()
}

// Creates an AND expression for all the tag names specified
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
func TagNames(expression Expression) []string {
	names := make([]string, 0, 10)
	names = tagNames(expression, names)

	return names
}

// Retrieves the set of value names from an expression
func ValueNames(expression Expression) []string {
	names := make([]string, 0, 10)
	names = valueNames(expression, names)

	return names
}

// unexported

func tagNames(expression Expression, names []string) []string {
	switch exp := expression.(type) {
	case TagExpression:
		names = append(names, exp.Name)
	case NotExpression:
		names = tagNames(exp.Operand, names)
	case AndExpression:
		names = tagNames(exp.LeftOperand, names)
		names = tagNames(exp.RightOperand, names)
	case OrExpression:
		names = tagNames(exp.LeftOperand, names)
		names = tagNames(exp.RightOperand, names)
	case ComparisonExpression:
		names = append(names, exp.Tag.Name)
	default:
		panic("unsupported token type")
	}

	return names
}

func valueNames(expression Expression, names []string) []string {
	switch exp := expression.(type) {
	case TagExpression:
		// nowt
	case NotExpression:
		names = valueNames(exp.Operand, names)
	case AndExpression:
		names = valueNames(exp.LeftOperand, names)
		names = valueNames(exp.RightOperand, names)
	case OrExpression:
		names = valueNames(exp.LeftOperand, names)
		names = valueNames(exp.RightOperand, names)
	case ComparisonExpression:
		names = append(names, exp.Value.Name)
	default:
		panic("unsupported token type")
	}

	return names
}
