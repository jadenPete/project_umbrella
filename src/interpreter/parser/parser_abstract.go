package parser

import (
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/parser/parser_types"
)

type Expression interface {
	Position() *errors.Position
}

type Assignment struct {
	Names []*Identifier
	Value Expression
}

func (assignment *Assignment) Position() *errors.Position {
	return &errors.Position{
		Start: assignment.Names[0].Position().Start,
		End:   assignment.Value.Position().End,
	}
}

type ExpressionList struct {
	Children []Expression
}

func (expressionList *ExpressionList) Position() *errors.Position {
	return &errors.Position{
		Start: expressionList.Children[0].Position().Start,
		End:   expressionList.Children[len(expressionList.Children)-1].Position().End,
	}
}

type Call struct {
	Function  Expression
	Arguments []Expression
	position  *errors.Position
}

func (call *Call) Position() *errors.Position {
	return call.position
}

type Float struct {
	Value    float64
	position *errors.Position
}

func (float *Float) Position() *errors.Position {
	return float.position
}

type Function struct {
	Name       *Identifier
	Parameters []*Identifier
	Value      *ExpressionList
	position   *errors.Position
}

func (function *Function) Position() *errors.Position {
	return function.position
}

type Identifier struct {
	Value    string
	position *errors.Position
}

func (identifier *Identifier) Position() *errors.Position {
	return identifier.position
}

type Integer struct {
	Value    int64
	position *errors.Position
}

func (integer *Integer) Position() *errors.Position {
	return integer.position
}

type Select struct {
	Value Expression
	Field *Identifier
	Type  parser_types.SelectType
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

func (string_ *String) Position() *errors.Position {
	return string_.position
}
