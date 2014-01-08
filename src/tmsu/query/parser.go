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
)

type Parser struct {
	scanner *Scanner
}

func NewParser(scanner *Scanner) Parser {
	return Parser{scanner}
}

func (parser Parser) Parse() (Expression, error) {
	return parser.expression()
}

type Expression interface {
}

type EmptyExpression struct {
}

type OrExpression struct {
	LeftOperand  Expression
	RightOperand Expression
}

type AndExpression struct {
	LeftOperand  Expression
	RightOperand Expression
}

type EqualsExpression struct {
	Tag   TagExpression
	Value ValueExpression
}

type NotExpression struct {
	Operand Expression
}

type TagExpression struct {
	Name string
}

type ValueExpression struct {
	Name string
}

// unexported

func (parser Parser) expression() (Expression, error) {
	expression, err := parser.or()
	if err != nil {
		return nil, err
	}

	token, err := parser.scanner.LookAhead()
	switch token.(type) {
	case EndToken:
		return expression, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v.", Type(token))
	}
}

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
		case EndToken, CloseParenToken:
			return leftOperand, nil
		default:
			return nil, fmt.Errorf("unexpected token: %v.", Type(token))
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
		case OrOperatorToken, CloseParenToken, EndToken:
			return leftOperand, nil
		case NotOperatorToken, SymbolToken, OpenParenToken:
			rightOperand, err := parser.not()
			if err != nil {
				return nil, err
			}

			leftOperand = AndExpression{leftOperand, rightOperand}
		default:
			return nil, fmt.Errorf("unexpected token: %v.", Type(token))
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

		operand, err := parser.not()
		if err != nil {
			return nil, err
		}

		return NotExpression{operand}, nil
	case OpenParenToken:
		parser.scanner.Next()

		operand, err := parser.or()
		if err != nil {
			return nil, err
		}

		token2, err := parser.scanner.Next()
		if err != nil {
			return nil, err
		}

		switch token2.(type) {
		case CloseParenToken:
			return operand, nil
		default:
			return nil, fmt.Errorf("unexpected token: %v", Type(token2))
		}
	case SymbolToken:
		operand, err := parser.equals()
		if err != nil {
			return nil, err
		}

		return operand, nil
	default:
		return nil, fmt.Errorf("unexpected token: %v.", Type(token))
	}
}

func (parser Parser) equals() (Expression, error) {
	tag, err := parser.tag()
	if err != nil {
		return nil, err
	}

	token, err := parser.scanner.LookAhead()
	if err != nil {
		return nil, err
	}

	switch token.(type) {
	case EqualOperatorToken:
		parser.scanner.Next()

		value, err := parser.value()
		if err != nil {
			return nil, err
		}

		return EqualsExpression{tag, value}, nil
	}

	return tag, nil
}

func (parser Parser) tag() (TagExpression, error) {
	token, err := parser.scanner.Next()
	if err != nil {
		return TagExpression{}, err
	}

	switch typedToken := token.(type) {
	case SymbolToken:
		return TagExpression{typedToken.name}, nil
	default:
		return TagExpression{}, fmt.Errorf("unexpected token: %v.", Type(token))
	}
}

func (parser Parser) value() (ValueExpression, error) {
	token, err := parser.scanner.Next()
	if err != nil {
		return ValueExpression{}, err
	}

	switch typedToken := token.(type) {
	case SymbolToken:
		return ValueExpression{typedToken.name}, nil
	default:
		return ValueExpression{}, fmt.Errorf("unexpected token: %v", Type(token))
	}
}
