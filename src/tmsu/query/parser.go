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
	"reflect"
)

type Parser struct {
	scanner *Scanner
}

func NewParser(scanner *Scanner) Parser {
	return Parser{scanner}
}

func (parser Parser) Parse() (Expression, error) {
	return parser.or()
}

type Expression interface {
}

type OrExpression struct {
	LeftOperand  Expression
	RightOperand Expression
}

type AndExpression struct {
	LeftOperand  Expression
	RightOperand Expression
}

type NotExpression struct {
	Operand Expression
}

type TagExpression struct {
	Name string
}

// unexported

func (parser Parser) or() (Expression, error) {
	leftOperand, err := parser.and()
	if err != nil {
		return OrExpression{}, err
	}

	for {
		token, err := parser.scanner.LookAhead()
		if err != nil {
			return nil, err
		}

		switch token.(type) {
		case OrOperatorToken:
			parser.scanner.Next()
			rightOperand, err := parser.and()
			if err != nil {
				return nil, err
			}

			leftOperand = OrExpression{leftOperand, rightOperand}
		case EndToken:
			return leftOperand, nil
		default:
			return nil, fmt.Errorf("unexpected token '%v': expecting 'or'.", token)
		}
	}
}

func (parser Parser) and() (Expression, error) {
	leftOperand, err := parser.not()
	if err != nil {
		return nil, err
	}

	for {
		token, err := parser.scanner.LookAhead()
		if err != nil {
			return nil, err
		}

		switch token.(type) {
		case AndOperatorToken:
			parser.scanner.Next()
			rightOperand, err := parser.not()
			if err != nil {
				return nil, err
			}

			leftOperand = AndExpression{leftOperand, rightOperand}
		case OrOperatorToken, EndToken:
			return leftOperand, nil
		case NotOperatorToken, TagToken:
			rightOperand, err := parser.not()
			if err != nil {
				return nil, err
			}

			leftOperand = AndExpression{leftOperand, rightOperand}
		default:
			return nil, fmt.Errorf("unexpected token '%v': expecting 'and' or tag.", token)
		}
	}
}

func (parser Parser) not() (Expression, error) {
	token, err := parser.scanner.LookAhead()
	if err != nil {
		return nil, err
	}

	switch token.(type) {
	case NotOperatorToken:
		parser.scanner.Next()

		operand, err := parser.tag()
		if err != nil {
			return nil, err
		}

		return NotExpression{operand}, nil
	default:
		operand, err := parser.tag()
		if err != nil {
			return nil, err
		}

		return operand, nil
	}
}

func (parser Parser) tag() (Expression, error) {
	token, err := parser.scanner.Next()
	if err != nil {
		return nil, err
	}

	switch typedToken := token.(type) {
	case TagToken:
		return TagExpression{typedToken.name}, nil
	default:
		return nil, fmt.Errorf("unexpected token '%v'. Expected tag.", reflect.TypeOf(token).String())
	}
}
