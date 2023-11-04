package parser

import (
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/parser/parser_types"
)

type Expression interface {
	Children() []Expression
	Position() *errors.Position
}

type Assignment struct {
	Names []*Identifier
	Value Expression
}

func (assignment *Assignment) Children() []Expression {
	result := make([]Expression, 0, len(assignment.Names)+1)

	for _, name := range assignment.Names {
		result = append(result, name)
	}

	result = append(result, assignment.Value)

	return result
}

func (assignment *Assignment) Position() *errors.Position {
	return &errors.Position{
		Start: assignment.Names[0].Position().Start,
		End:   assignment.Value.Position().End,
	}
}

type ExpressionList struct {
	Children_ []Expression
}

func (expressionList *ExpressionList) Children() []Expression {
	return expressionList.Children_
}

func (expressionList *ExpressionList) Position() *errors.Position {
	if len(expressionList.Children_) == 0 {
		return nil
	}

	return &errors.Position{
		Start: expressionList.Children_[0].Position().Start,
		End:   expressionList.Children_[len(expressionList.Children_)-1].Position().End,
	}
}

type Call struct {
	Function  Expression
	Arguments []Expression
	position  *errors.Position
}

func (call *Call) Children() []Expression {
	return append([]Expression{call.Function}, call.Arguments...)
}

func (call *Call) Position() *errors.Position {
	return call.position
}

type Float struct {
	Value    float64
	position *errors.Position
}

func (*Float) Children() []Expression {
	return []Expression{}
}

func (float *Float) Position() *errors.Position {
	return float.position
}

type Function struct {
	Name       *Identifier
	Parameters []*Identifier
	Body       *ExpressionList
	position   *errors.Position
}

func (function *Function) Children() []Expression {
	result := make([]Expression, 0, len(function.Parameters)+2)

	if function.Name != nil {
		result = append(result, function.Name)
	}

	for _, parameter := range function.Parameters {
		result = append(result, parameter)
	}

	result = append(result, function.Body)

	return result
}

func (function *Function) Position() *errors.Position {
	return function.position
}

type Identifier struct {
	Value    string
	position *errors.Position
}

func (*Identifier) Children() []Expression {
	return []Expression{}
}

func (identifier *Identifier) Position() *errors.Position {
	return identifier.position
}

type Integer struct {
	Value    int64
	position *errors.Position
}

func (*Integer) Children() []Expression {
	return []Expression{}
}

func (integer *Integer) Position() *errors.Position {
	return integer.position
}

type Select struct {
	Value Expression
	Field *Identifier
	Type  parser_types.SelectType
}

func (select_ *Select) Children() []Expression {
	return []Expression{select_.Value, select_.Field}
}

func (select_ *Select) Position() *errors.Position {
	return &errors.Position{
		Start: select_.Value.Position().Start,
		End:   select_.Field.Position().End,
	}
}

type String struct {
	Value    string
	position *errors.Position
}

func (*String) Children() []Expression {
	return []Expression{}
}

func (string_ *String) Position() *errors.Position {
	return string_.position
}
