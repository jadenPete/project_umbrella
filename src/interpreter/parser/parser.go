package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
)

// Union expressions

type ConcreteStatement interface {
	Abstract() Expression
	Tokens_() []lexer.Token

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
		rightHandSideAbstract := rightHandSide.Operand().Abstract()

		result = &Call{
			Function: &Select{
				Value:   result,
				Field:   rightHandSide.Operator(),
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
	Operator() *Identifier
	Operand() Operand
}

type ConcretePrimary interface {
	Abstract() Expression
	primary()
}

// Statements

type ConcreteAssignment struct {
	Name   *ConcreteIdentifier `parser:"@@ (IndentToken | OutdentToken | NewlineToken)* '=' (IndentToken | OutdentToken | NewlineToken)*"`
	Tail   *ConcreteAssignment `parser:"  (@@"`
	Value  ConcreteExpression  `parser:" | @@)"`
	Tokens []lexer.Token
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

func (concrete *ConcreteAssignment) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteAssignment) concreteStatement() {}

type ConcreteFunction struct {
	Declaration   *ConcreteFunctionDeclaration   `parser:"@@ ':' NewlineToken"`
	StatementList *ConcreteIndentedStatementList `parser:"@@"`
	Tokens        []lexer.Token
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

func (concrete *ConcreteFunction) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteFunction) concreteStatement() {}

type ConcreteFunctionDeclaration struct {
	Name       *ConcreteIdentifier         `parser:"'fn' (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)* '(' (IndentToken | OutdentToken | NewlineToken)*"`
	Parameters *ConcreteFunctionParameters `parser:"@@? (IndentToken | OutdentToken | NewlineToken)* ')'"`
	Pos        lexer.Position
}

type ConcreteFunctionParameters struct {
	Head *ConcreteIdentifier         `parser:"@@"`
	Tail *ConcreteFunctionParameters `parser:"((IndentToken | OutdentToken | NewlineToken)* ',' (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type ConcreteIndentedStatementList struct {
	Value *ConcreteStatementList `parser:"NewlineToken* (IndentToken @@ (OutdentToken | EOF))?"`
}

func (concrete *ConcreteIndentedStatementList) Abstract() Expression {
	return concrete.AbstractExpressionList()
}

func (concrete *ConcreteIndentedStatementList) AbstractExpressionList() *ExpressionList {
	if concrete.Value == nil {
		return &ExpressionList{
			Children: []Expression{},
		}
	}

	return concrete.Value.AbstractExpressionList()
}

type ConcreteStatementList struct {
	Children []ConcreteStatement `parser:"NewlineToken* (@@ (NewlineToken+ @@)* NewlineToken*)?"`
}

/*
 * This function is responsible for parsing statement lists into concrete syntax trees comprising
 * zero or more statements. Statement lists cannot be parsed with participle's tax syntax
 * (https://github.com/alecthomas/participle#tag-syntax) because they aren't technically
 * context-free.
 *
 * Because Krait's syntax adheres to the off-side rule
 * (https://en.wikipedia.org/wiki/Off-side_rule), it suffers from the problem of what I've deeme
 * hanging expressions. Consider the following program.
 *
 * ```
 * message =\n
 * ->"Hello, world!"<-\n
 * \n
 * println(message)\n
 * ```
 *
 * Although the assignment to `message` and call to `println` will be parsed correctly,
 * the former is implicitly followed by an outdent, then newline token. This is a problem because
 * conjoining both statements into a final statement list requires they be separated only by
 * newlines.
 *
 * To solve this issue, we "float" indentation parsed in an expression out and to the right of that
 * expression, cancelling it out with converse indentation where possible. In the above example, the
 * indent token inside the assignment would float outward and cancel with the outdent following the
 * assignment, allowing for the formation of a proper statement list.
 */
func (concrete *ConcreteStatementList) Parse(lexer_ *lexer.PeekingLexer) error {
	statementParser := (*participle.Parser[ConcreteStatement])(parser)
	consumeAndAppendStatement := func() bool {
		statementPointer, err := statementParser.ParseFromLexer(
			lexer_,
			participle.AllowTrailing(true),
		)

		if err != nil {
			return false
		}

		statement := *statementPointer

		netIndentation := 0

		for _, token := range statement.Tokens_() {
			if token.Type == IndentToken {
				netIndentation++
			} else if token.Type == OutdentToken {
				netIndentation--
			}
		}

		indentationTokenNeeded := IndentToken

		if netIndentation > 0 {
			indentationTokenNeeded = OutdentToken
		}

		for i := 0; i < common.Abs(netIndentation); i++ {
			if lexer_.Peek().Type != indentationTokenNeeded {
				return false
			}

			lexer_.Next()
		}

		if err == nil {
			concrete.Children = append(concrete.Children, statement)

			return true
		}

		return false
	}

	consumeNewlines := func() {
		for lexer_.Peek().Type == NewlineToken {
			lexer_.Next()
		}
	}

	consumeAndAppendAdditionalStatement := func() bool {
		if token := lexer_.Next(); token.Type != NewlineToken {
			return false
		}

		consumeNewlines()

		return consumeAndAppendStatement()
	}

	consumeNewlines()

	concrete.Children = make([]ConcreteStatement, 0)

	if consumeAndAppendStatement() {
		for {
			checkpoint := lexer_.MakeCheckpoint()

			if !consumeAndAppendAdditionalStatement() {
				lexer_.LoadCheckpoint(checkpoint)

				break
			}
		}

		consumeNewlines()
	}

	return nil
}

func (concrete *ConcreteStatementList) Abstract() Expression {
	return concrete.AbstractExpressionList()
}

func (concrete *ConcreteStatementList) AbstractExpressionList() *ExpressionList {
	children := make([]Expression, 0)

	for _, child := range concrete.Children {
		children = append(children, child.Abstract())
	}

	return &ExpressionList{
		Children: children,
	}
}

// Multi-token expressions

type ConcreteInfixMiscellaneous struct {
	Left_  *ConcreteInfixAddition             `parser:"@@"`
	Right_ []*ConcreteInfixMiscellaneousRight `parser:"@@*"`
	Tokens []lexer.Token
}

func (concrete *ConcreteInfixMiscellaneous) Abstract() Expression {
	return AbstractInfixOperation(
		ConcreteInfixOperation[*ConcreteInfixAddition, *ConcreteInfixMiscellaneousRight](concrete),
	)
}

func (concrete *ConcreteInfixMiscellaneous) Left() *ConcreteInfixAddition {
	return concrete.Left_
}

func (concrete *ConcreteInfixMiscellaneous) Right() []*ConcreteInfixMiscellaneousRight {
	return concrete.Right_
}

func (concrete *ConcreteInfixMiscellaneous) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteInfixMiscellaneous) concreteStatement()  {}
func (*ConcreteInfixMiscellaneous) concreteExpression() {}

type ConcreteInfixMiscellaneousRight struct {
	OperatorOne *ConcreteIdentifier    `parser:"(((IndentToken | OutdentToken)* @@ (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *ConcreteIdentifier    `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken)*))"`
	Operand_    *ConcreteInfixAddition `parser:"@@"`
}

func (concrete *ConcreteInfixMiscellaneousRight) Operator() *Identifier {
	if concrete.OperatorOne == nil {
		return concrete.OperatorTwo.AbstractIdentifier()
	}

	return concrete.OperatorOne.AbstractIdentifier()
}

func (concrete *ConcreteInfixMiscellaneousRight) Operand() *ConcreteInfixAddition {
	return concrete.Operand_
}

type ConcreteInfixAddition struct {
	Left_  *ConcreteInfixMultiplication  `parser:"@@"`
	Right_ []*ConcreteInfixAdditionRight `parser:"@@*"`
}

func (concrete *ConcreteInfixAddition) Abstract() Expression {
	return AbstractInfixOperation(
		ConcreteInfixOperation[*ConcreteInfixMultiplication, *ConcreteInfixAdditionRight](concrete),
	)
}

func (concrete *ConcreteInfixAddition) Left() *ConcreteInfixMultiplication {
	return concrete.Left_
}

func (concrete *ConcreteInfixAddition) Right() []*ConcreteInfixAdditionRight {
	return concrete.Right_
}

func (*ConcreteInfixAddition) concreteInfixOperand() {}

type ConcreteInfixAdditionRight struct {
	OperatorOne *ConcreteInfixAdditionOperator `parser:"(((IndentToken | OutdentToken)* @@ (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *ConcreteInfixAdditionOperator `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken)*))"`
	Operand_    *ConcreteInfixMultiplication   `parser:"@@"`
}

func (concrete *ConcreteInfixAdditionRight) Operator() *Identifier {
	if concrete.OperatorOne == nil {
		return concrete.OperatorTwo.AbstractIdentifier()
	}

	return concrete.OperatorOne.AbstractIdentifier()

}

func (concrete *ConcreteInfixAdditionRight) Operand() *ConcreteInfixMultiplication {
	return concrete.Operand_
}

type ConcreteInfixAdditionOperator struct {
	Identifier string `parser:"@('+' | '-')"`
	Tokens     []lexer.Token
}

func (concrete *ConcreteInfixAdditionOperator) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Identifier,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

type ConcreteInfixMultiplication struct {
	Left_  *ConcreteCall                       `parser:"@@"`
	Right_ []*ConcreteInfixMultiplicationRight `parser:"@@*"`
}

func (concrete *ConcreteInfixMultiplication) Abstract() Expression {
	return AbstractInfixOperation(
		ConcreteInfixOperation[*ConcreteCall, *ConcreteInfixMultiplicationRight](concrete),
	)
}

func (concrete *ConcreteInfixMultiplication) Left() *ConcreteCall {
	return concrete.Left_
}

func (concrete *ConcreteInfixMultiplication) Right() []*ConcreteInfixMultiplicationRight {
	return concrete.Right_
}

func (*ConcreteInfixMultiplication) concreteInfixOperand() {}

type ConcreteInfixMultiplicationRight struct {
	OperatorOne *ConcreteInfixMultiplicationOperator `parser:"(((IndentToken | OutdentToken)* @@ (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *ConcreteInfixMultiplicationOperator `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken)*))"`
	Operand_    *ConcreteCall                        `parser:"@@"`
}

func (concrete *ConcreteInfixMultiplicationRight) Operator() *Identifier {
	if concrete.OperatorOne == nil {
		return concrete.OperatorTwo.AbstractIdentifier()
	}

	return concrete.OperatorOne.AbstractIdentifier()
}

func (concrete *ConcreteInfixMultiplicationRight) Operand() *ConcreteCall {
	return concrete.Operand_
}

type ConcreteInfixMultiplicationOperator struct {
	Identifier string `parser:"@('*' | '/')"`
	Tokens     []lexer.Token
}

func (concrete *ConcreteInfixMultiplicationOperator) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Identifier,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

type ConcreteCall struct {
	Left  *ConcreteSelect      `parser:"@@"`
	Right []*ConcreteCallRight `parser:"@@*"`
}

func (concrete *ConcreteCall) Abstract() Expression {
	result := concrete.Left.Abstract()

	for _, rightHandSide := range concrete.Right {
		if rightHandSide.Select == nil {
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
		} else {
			result = &Select{
				Value:   result,
				Field:   rightHandSide.Select.Field.AbstractIdentifier(),
				IsInfix: false,
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
	Field *ConcreteIdentifier `parser:"(IndentToken | OutdentToken | NewlineToken)* '.' (IndentToken | OutdentToken | NewlineToken)* @@"`
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

var parser = participle.MustBuild[ConcreteStatementList](
	participle.Lexer(&LexerDefinition{}),
	participle.Union[ConcreteStatement](
		&ConcreteAssignment{},
		&ConcreteFunction{},
		&ConcreteInfixMiscellaneous{},
	),

	participle.Union[ConcreteExpression](&ConcreteInfixMiscellaneous{}),
	participle.Union[ConcretePrimary](
		&ConcreteFloat{},
		&ConcreteIdentifier{},
		&ConcreteInteger{},
		&ConcreteString{},
	),

	participle.Unquote("StringToken"),
	participle.UseLookahead(participle.MaxLookahead),
)

func ParseString(source string) (*ConcreteStatementList, error) {
	return parser.ParseString("", source)
}

func tokenSyntaxTreePosition(token *lexer.Token) *errors.Position {
	return &errors.Position{
		Start: token.Pos.Offset,
		End:   token.Pos.Offset + len(token.Value),
	}
}
