package main

import (
	"github.com/alecthomas/participle/v2/lexer"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
)

// Union expressions

type ConcreteStatement interface {
	Abstract() Expression
	concreteStatement()
}

type ConcreteExpression interface {
	ConcreteStatement

	concreteExpression()
}

type ConcreteInfixOperation[
	Operand ConcreteInfixOperand,
	Right ConcreteInfixOperationRight[Operand],
] interface {
	Left() Operand
	Right() []Right
}

func AbstractInfixOperation[
	Operand ConcreteInfixOperand,
	Right ConcreteInfixOperationRight[Operand],
](
	concrete ConcreteInfixOperation[Operand, Right],
) Expression {
	result := concrete.Left().Abstract()

	for _, rightHandSide := range concrete.Right() {
		operator := rightHandSide.OperatorOne()

		if operator == nil {
			operator = rightHandSide.OperatorTwo()
		}

		rightHandSideAbstract := rightHandSide.Operand().Abstract()

		result = &Call{
			Function: &Select{
				Value:   result,
				Field:   operator.AbstractIdentifier(),
				IsInfix: true,
			},

			Arguments: []Expression{rightHandSideAbstract},
			position: &errors.Position{
				Start: result.Position().Start,
				End:   rightHandSideAbstract.Position().End,
			},
		}
	}

	return result
}

type ConcreteInfixOperand interface {
	Abstract() Expression
	concreteInfixOperand()
}

type ConcreteInfixOperationRight[Operand ConcreteInfixOperand] interface {
	OperatorOne() *ConcreteIdentifier
	OperatorTwo() *ConcreteIdentifier
	Operand() Operand
}

type ConcretePrimary interface {
	Abstract() Expression
	primary()
}

// Statements

type ConcreteAssignment struct {
	Name  *ConcreteIdentifier `parser:"@IdentifierToken (IndentToken | OutdentToken | NewlineToken)* '=' (IndentToken | OutdentToken | NewlineToken)*"`
	Value ConcreteExpression  `parser:" (@@"`
	Tail  *ConcreteAssignment `parser:"| @@)"`
}

func (concrete *ConcreteAssignment) Abstract() Expression {
	return concrete.AbstractAssignment()
}

func (concrete *ConcreteAssignment) AbstractAssignment() *Assignment {
	names, last := common.LinkedListToSlice[ConcreteAssignment, *Identifier](
		concrete,
		func(child *ConcreteAssignment) *Identifier { return child.Name.AbstractIdentifier() },
		func(child *ConcreteAssignment) *ConcreteAssignment { return child.Tail },
	)

	return &Assignment{
		Names: names,
		Value: last.Value.Abstract(),
	}
}

func (*ConcreteAssignment) concreteStatement() {}

type ConcreteFunction struct {
	Declaration   *ConcreteFunctionDeclaration   `parser:"@@ ':' NewlineToken"`
	StatementList *ConcreteIndentedStatementList `parser:"@@"`
}

func (concrete *ConcreteFunction) Abstract() Expression {
	return concrete.AbstractFunction()
}

func (concrete *ConcreteFunction) AbstractFunction() *Function {
	parameters, _ := common.LinkedListToSlice[ConcreteFunctionParameters, *Identifier](
		concrete.Declaration.Parameters,
		func(child *ConcreteFunctionParameters) *Identifier {
			return child.Head.AbstractIdentifier()
		},

		func(child *ConcreteFunctionParameters) *ConcreteFunctionParameters {
			return child.Tail
		},
	)

	value := concrete.StatementList.AbstractExpressionList()

	return &Function{
		Name:       concrete.Declaration.Name.AbstractIdentifier(),
		Parameters: parameters,
		Value:      value,
		position: &errors.Position{
			Start: concrete.Declaration.Pos.Offset,
			End:   value.Position().End,
		},
	}
}

func (*ConcreteFunction) concreteStatement() {}

type ConcreteFunctionDeclaration struct {
	Name       *ConcreteIdentifier         `parser:"'fn' (IndentToken | OutdentToken | NewlineToken)* @IdentifierToken (IndentToken | OutdentToken | NewlineToken)* '(' (IndentToken | OutdentToken | NewlineToken)*"`
	Parameters *ConcreteFunctionParameters `parser:"@@? (IndentToken | OutdentToken | NewlineToken)* ')'"`
	Pos        lexer.Position
}

type ConcreteFunctionParameters struct {
	Head *ConcreteIdentifier         `parser:"@IdentifierToken"`
	Tail *ConcreteFunctionParameters `parser:"((IndentToken | OutdentToken | NewlineToken)* ',' (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type ConcreteIndentedStatementList struct {
	Value *ConcreteStatementList `parser:"IndentToken @@ (OutdentToken | EOF)"`
}

func (concrete *ConcreteIndentedStatementList) Abstract() Expression {
	return concrete.AbstractExpressionList()
}

func (concrete *ConcreteIndentedStatementList) AbstractExpressionList() *ExpressionList {
	return concrete.Value.AbstractExpressionList()
}

type ConcreteStatementList struct {
	Head ConcreteStatement      `parser:"NewlineToken* @@"`
	Tail *ConcreteStatementList `parser:"NewlineToken @@ | NewlineToken*"`
}

func (concrete *ConcreteStatementList) Abstract() Expression {
	return concrete.AbstractExpressionList()
}

func (concrete *ConcreteStatementList) AbstractExpressionList() *ExpressionList {
	result, _ := common.LinkedListToSlice[ConcreteStatementList, Expression](
		concrete,
		func(child *ConcreteStatementList) Expression { return child.Head.Abstract() },
		func(child *ConcreteStatementList) *ConcreteStatementList { return child.Tail },
	)

	return &ExpressionList{
		Children: result,
	}
}

// Multi-token expressions

type ConcreteInfixAddition struct {
	left  *ConcreteInfixMultiplication  `parser:"@@"`
	right []*ConcreteInfixAdditionRight `parser:"@@*"`
}

func (concrete *ConcreteInfixAddition) Abstract() Expression {
	return AbstractInfixOperation(concrete)
}

func (concrete *ConcreteInfixAddition) Left() *ConcreteInfixMultiplication {
	return concrete.left
}

func (concrete *ConcreteInfixAddition) Right() []*ConcreteInfixAdditionRight {
	return concrete.right
}

func (*ConcreteInfixAddition) concreteStatement()    {}
func (*ConcreteInfixAddition) concreteExpression()   {}
func (*ConcreteInfixAddition) concreteInfixOperand() {}

type ConcreteInfixAdditionRight struct {
	operatorOne *ConcreteIdentifier          `parser:"(((IndentToken | OutdentToken)* @('+' | '-') (IndentToken | OutdentToken | NewlineToken)*)"`
	operatorTwo *ConcreteIdentifier          `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @('+' | '-') (IndentToken | OutdentToken)*))"`
	operand     *ConcreteInfixMultiplication `parser:"@@"`
}

func (concrete *ConcreteInfixAdditionRight) OperatorOne() *ConcreteIdentifier {
	return concrete.operatorOne
}

func (concrete *ConcreteInfixAdditionRight) OperatorTwo() *ConcreteIdentifier {
	return concrete.operatorTwo
}

func (concrete *ConcreteInfixAdditionRight) Operand() *ConcreteInfixMultiplication {
	return concrete.operand
}

type ConcreteInfixMultiplication struct {
	left  *ConcreteInfixMiscellaneous         `parser:"@@"`
	right []*ConcreteInfixMultiplicationRight `parser:"@@*"`
}

func (concrete *ConcreteInfixMultiplication) Abstract() Expression {
	return AbstractInfixOperation(concrete)
}

func (concrete *ConcreteInfixMultiplication) Left() *ConcreteInfixMiscellaneous {
	return concrete.left
}

func (concrete *ConcreteInfixMultiplication) Right() []*ConcreteInfixMultiplicationRight {
	return concrete.right
}

func (*ConcreteInfixMultiplication) concreteInfixOperand() {}

type ConcreteInfixMultiplicationRight struct {
	operatorOne *ConcreteIdentifier         `parser:"(((IndentToken | OutdentToken)* @('*' | '/') (IndentToken | OutdentToken | NewlineToken)*)"`
	operatorTwo *ConcreteIdentifier         `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @('*' | '/') (IndentToken | OutdentToken)*))"`
	operand     *ConcreteInfixMiscellaneous `parser:"@@"`
}

func (concrete *ConcreteInfixMultiplicationRight) OperatorOne() *ConcreteIdentifier {
	return concrete.operatorOne
}

func (concrete *ConcreteInfixMultiplicationRight) OperatorTwo() *ConcreteIdentifier {
	return concrete.operatorTwo
}

func (concrete *ConcreteInfixMultiplicationRight) Operand() *ConcreteInfixMiscellaneous {
	return concrete.operand
}

type ConcreteInfixMiscellaneous struct {
	left  *ConcreteCall                      `parser:"@@"`
	right []*ConcreteInfixMiscellaneousRight `parser:"@@*"`
}

func (concrete *ConcreteInfixMiscellaneous) Abstract() Expression {
	return AbstractInfixOperation(concrete)
}

func (concrete *ConcreteInfixMiscellaneous) Left() *ConcreteCall {
	return concrete.left
}

func (concrete *ConcreteInfixMiscellaneous) Right() []*ConcreteInfixMiscellaneousRight {
	return concrete.right
}

func (*ConcreteInfixMiscellaneous) concreteInfixOperand() {}

type ConcreteInfixMiscellaneousRight struct {
	operatorOne *ConcreteIdentifier `parser:"(((IndentToken | OutdentToken)* @IdentifierToken (IndentToken | OutdentToken | NewlineToken)*)"`
	operatorTwo *ConcreteIdentifier `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @IdentifierToken (IndentToken | OutdentToken)*))"`
	operand     *ConcreteCall       `parser:"@@"`
}

func (concrete *ConcreteInfixMiscellaneousRight) OperatorOne() *ConcreteIdentifier {
	return concrete.operatorOne
}

func (concrete *ConcreteInfixMiscellaneousRight) OperatorTwo() *ConcreteIdentifier {
	return concrete.operatorTwo
}

func (concrete *ConcreteInfixMiscellaneousRight) Operand() *ConcreteCall {
	return concrete.operand
}

type ConcreteCall struct {
	Left  *ConcreteSelect      `parser:"@@"`
	Right []*ConcreteCallRight `parser:"@@*"`
}

func (concrete *ConcreteCall) Abstract() Expression {
	result := concrete.Left.Abstract()

	for _, rightHandSide := range concrete.Right {
		if rightHandSide.Arguments == nil {
			result = &Select{
				Value:   result,
				Field:   rightHandSide.Select.Field.AbstractIdentifier(),
				IsInfix: false,
			}
		} else {
			arguments, _ := common.LinkedListToSlice[ConcreteCallArguments, Expression](
				rightHandSide.Arguments,
				func(child *ConcreteCallArguments) Expression { return child.Head.Abstract() },
				func(child *ConcreteCallArguments) *ConcreteCallArguments { return child.Tail },
			)

			lastToken := rightHandSide.Tokens[len(rightHandSide.Tokens)-1]

			result = &Call{
				Function:  result,
				Arguments: arguments,
				position: &errors.Position{
					Start: result.Position().Start,
					End:   lastToken.Pos.Offset + len(lastToken.Value),
				},
			}
		}
	}

	return result
}

func (*ConcreteCall) concreteInfixOperand() {}

type ConcreteCallRight struct {
	Arguments *ConcreteCallArguments `parser:"(IndentToken | OutdentToken)* '(' (IndentToken | OutdentToken | NewlineToken)* @@? (IndentToken | OutdentToken | NewlineToken)* ')'"`
	Select    *ConcreteSelectRight   `parser:"| @@"`
	Tokens    []lexer.Token
}

type ConcreteCallArguments struct {
	Head ConcreteExpression     `parser:"@@"`
	Tail *ConcreteCallArguments `parser:" ((IndentToken | OutdentToken | NewlineToken)* ',' (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type ConcreteSelect struct {
	Left  ConcretePrimary        `parser:"@@"`
	Right []*ConcreteSelectRight `parser:"@@*"`
}

func (concrete *ConcreteSelect) Abstract() Expression {
	result := concrete.Left.Abstract()

	for _, rightHandSide := range concrete.Right {
		result = &Select{
			Value:   result,
			Field:   rightHandSide.Field.AbstractIdentifier(),
			IsInfix: false,
		}
	}

	return result
}

type ConcreteSelectRight struct {
	Field *ConcreteIdentifier `parser:"(IndentToken | OutdentToken | NewlineToken)* '.' (IndentToken | OutdentToken | NewlineToken)* @IdentifierToken"`
}

// Single-token expressions

type ConcreteFloat struct {
	Value  float64 `parser:"@FloatToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteFloat) Abstract() Expression {
	return concrete.AbstractFloat()
}

func (concrete *ConcreteFloat) AbstractFloat() *Float {
	return &Float{
		Value:    concrete.Value,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

func (*ConcreteFloat) primary() {}

type ConcreteIdentifier struct {
	Value  string `parser:"@IdentifierToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteIdentifier) Abstract() Expression {
	return concrete.AbstractIdentifier()
}

func (concrete *ConcreteIdentifier) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Value,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

/*
 * This method is necessary for `Identifier` to implement the `participle.Capturable` interface,
 * allowing specific identifier tokens to be matched into the `Identifier` terminal struct.
 *
 * Note that because of a bug in Participle, `Identifier` can't be captured with `@@`:
 * https://github.com/alecthomas/participle/issues/140
 */
func (identifier *ConcreteIdentifier) Capture(values []string) error {
	identifier.Value = values[0]

	return nil
}

func (*ConcreteIdentifier) primary() {}

type ConcreteInteger struct {
	Value  int64 `parser:"@IntegerToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteInteger) Abstract() Expression {
	return concrete.AbstractInteger()
}

func (concrete *ConcreteInteger) AbstractInteger() *Integer {
	return &Integer{
		Value:    concrete.Value,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

func (*ConcreteInteger) primary() {}

type ConcreteString struct {
	Value  string `parser:"@StringToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteString) Abstract() Expression {
	return concrete.AbstractString()
}

func (concrete *ConcreteString) AbstractString() *String {
	return &String{
		Value:    concrete.Value,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

func (*ConcreteString) primary() {}

func tokenSyntaxTreePosition(token *lexer.Token) *errors.Position {
	return &errors.Position{
		Start: token.Pos.Offset,
		End:   token.Pos.Offset + len(token.Value),
	}
}
