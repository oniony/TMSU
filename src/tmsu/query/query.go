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

func Parse(query string) (Expression, error) {
	scanner := NewScanner(query)
	parser := NewParser(scanner)

	return parser.Parse()
}

func HasAll(tagNames []string) Expression {
	var expression Expression = EmptyExpression{}

	for _, tagName := range tagNames {
		expression = AndExpression{expression, TagExpression{tagName}}
	}

	return expression
}

func TagNames(expression Expression) []string {
	names := make([]string, 0, 10)
	names = tagNames(expression, names)

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
	default:
		panic("unsupported token type")
	}

	return names
}
