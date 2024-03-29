package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/parser/parser_types"
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
		abstractRightHandSide := rightHandSide.Operand().Abstract()

		result = &Call{
			Function: &Select{
				Value: result,
				Field: rightHandSide.Operator(),
				Type:  parser_types.InfixSelect,
			},

			Arguments: []Expression{abstractRightHandSide},
			position: &errors.Position{
				Filename: result.Position().Filename,
				Start:    result.Position().Start,
				End:      abstractRightHandSide.Position().End,
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
	Name   *ConcreteIdentifier `parser:"@@ (IndentToken | OutdentToken | NewlineToken)* '=':AssignmentOperatorToken (IndentToken | OutdentToken | NewlineToken)*"`
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
		Names_: names,
		Value:  last.Value.Abstract(),
	}
}

func (concrete *ConcreteAssignment) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteAssignment) concreteStatement() {}

type ConcreteBlock struct {
	Expression    ConcreteExpression     `parser:"':':ColonToken (@@"`
	StatementList *ConcreteStatementList `parser:"              | NewlineToken+ IndentToken @@ (OutdentToken | EOF))?"`
}

func (concrete *ConcreteBlock) AbstractExpressionList() *ExpressionList {
	if concrete.Expression != nil {
		return &ExpressionList{
			Children_: []Expression{concrete.Expression.Abstract()},
		}
	}

	if concrete.StatementList != nil {
		return concrete.StatementList.AbstractExpressionList()
	}

	return &ExpressionList{
		Children_: []Expression{},
	}
}

type ConcreteFunction struct {
	Name              *ConcreteIdentifier                `parser:"'fn':FunctionKeywordToken (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)*"`
	ParametersAndBody *ConcreteFunctionParametersAndBody `parser:"@@"`
	Tokens            []lexer.Token
}

func (concrete *ConcreteFunction) Abstract() Expression {
	abstractParameters, abstractBody := concrete.ParametersAndBody.Abstract()

	return &Function{
		Name:       concrete.Name.AbstractIdentifier(),
		Parameters: abstractParameters,
		Body:       abstractBody,
		position:   tokenListSyntaxTreePosition(concrete.Tokens),
	}
}

func (concrete *ConcreteFunction) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteFunction) concreteStatement() {}

type ConcreteFunctionOrStructParameters struct {
	Head *ConcreteIdentifier                 `parser:"@@"`
	Tail *ConcreteFunctionOrStructParameters `parser:"((IndentToken | OutdentToken | NewlineToken)* ',':CommaToken (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

func AbstractFunctionOrStructParameters(concrete *ConcreteFunctionOrStructParameters) []*Identifier {
	result, _ := common.LinkedListToSlice[ConcreteFunctionOrStructParameters, *Identifier](
		concrete,
		func(child *ConcreteFunctionOrStructParameters) *Identifier {
			return child.Head.AbstractIdentifier()
		},

		func(child *ConcreteFunctionOrStructParameters) *ConcreteFunctionOrStructParameters {
			return child.Tail
		},
	)

	return result
}

type ConcreteFunctionParametersAndBody struct {
	Parameters *ConcreteFunctionOrStructParameters `parser:"'(':LeftParenthesisToken (IndentToken | OutdentToken | NewlineToken)* @@? (IndentToken | OutdentToken | NewlineToken)* ')':RightParenthesisToken NewlineToken*"`
	Body       *ConcreteBlock                      `parser:"@@"`
}

func (concrete *ConcreteFunctionParametersAndBody) Abstract() ([]*Identifier, *ExpressionList) {
	abstractParameters := AbstractFunctionOrStructParameters(concrete.Parameters)
	abstractBody := concrete.Body.AbstractExpressionList()

	return abstractParameters, abstractBody
}

type ConcreteStatementList struct {
	Children []ConcreteStatement
}

/*
 * This function is responsible for parsing statement lists into concrete syntax trees comprising
 * zero or more statements. Statement lists cannot be parsed with Participle's tax syntax
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

	concrete.Children = []ConcreteStatement{}

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

func (concrete *ConcreteStatementList) AbstractExpressionList() *ExpressionList {
	children := make([]Expression, 0, len(concrete.Children))

	for _, child := range concrete.Children {
		children = append(children, child.Abstract())
	}

	return &ExpressionList{
		Children_: children,
	}
}

type ConcreteStruct struct {
	Name       *ConcreteIdentifier                 `parser:"'struct':StructKeywordToken (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)* '(':LeftParenthesisToken (IndentToken | OutdentToken | NewlineToken)*"`
	Parameters *ConcreteFunctionOrStructParameters `parser:"@@ (IndentToken | OutdentToken | NewlineToken)* ')':RightParenthesisToken NewlineToken*"`
	Body       *ConcreteBlock                      `parser:"@@"`
	Tokens     []lexer.Token
}

func (concrete *ConcreteStruct) Abstract() Expression {
	abstractBody := concrete.Body.AbstractExpressionList()
	abstractParameters := AbstractFunctionOrStructParameters(concrete.Parameters.Tail)
	argumentFields := make([]Expression, 0, len(abstractParameters))
	nonArgumentFields := make([]Expression, 0, len(abstractParameters))

	addField := func(identifier *Identifier, isArgument bool) {
		field := AbstractTuple(
			[]Expression{
				&String{
					Value:    identifier.Value,
					position: nil,
				},

				identifier,
			},

			nil,
		)

		if isArgument {
			argumentFields = append(argumentFields, field)
		} else {
			nonArgumentFields = append(nonArgumentFields, field)
		}
	}

	for _, parameter := range abstractParameters {
		addField(parameter, true)
	}

	for _, statement := range abstractBody.Children() {
		if declaration, ok := statement.(Declaration); ok {
			for _, name := range declaration.Names() {
				addField(name, false)
			}
		}
	}

	fieldFactory := &Function{
		Name: nil,
		Parameters: []*Identifier{
			{
				Value:    "self",
				position: concrete.Parameters.Head.Abstract().Position(),
			},
		},

		Body: &ExpressionList{
			Children_: append(abstractBody.Children(), []Expression{
				AbstractTuple(nonArgumentFields, nil),
			}...),
		},

		position: nil,
	}

	resultName := concrete.Name.AbstractIdentifier()
	result := &Function{
		Name:       resultName,
		Parameters: abstractParameters,
		Body:       nil,
		position:   tokenListSyntaxTreePosition(concrete.Tokens),
	}

	result.Body = &ExpressionList{
		Children_: []Expression{
			&Call{
				Function: &Identifier{
					Value:    "__struct__",
					position: nil,
				},

				Arguments: []Expression{
					&String{
						Value:    concrete.Name.Value,
						position: nil,
					},

					resultName,
					fieldFactory,
					AbstractTuple(argumentFields, nil),
				},

				position: nil,
			},
		},
	}

	return result
}

func (concrete *ConcreteStruct) Tokens_() []lexer.Token {
	return concrete.Tokens
}

func (*ConcreteStruct) concreteStatement() {}

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
	OperatorOne *ConcreteOperator      `parser:"(((IndentToken | OutdentToken)* @@ (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *ConcreteOperator      `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken)*))"`
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
	Identifier string `parser:"@('+':OperatorToken | '-':OperatorToken)"`
	Tokens     []lexer.Token
}

func (concrete *ConcreteInfixAdditionOperator) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Identifier,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

type ConcreteInfixMultiplication struct {
	Left_  *ConcretePrefixOperation            `parser:"@@"`
	Right_ []*ConcreteInfixMultiplicationRight `parser:"@@*"`
}

func (concrete *ConcreteInfixMultiplication) Abstract() Expression {
	return AbstractInfixOperation(
		ConcreteInfixOperation[*ConcretePrefixOperation, *ConcreteInfixMultiplicationRight](concrete),
	)
}

func (concrete *ConcreteInfixMultiplication) Left() *ConcretePrefixOperation {
	return concrete.Left_
}

func (concrete *ConcreteInfixMultiplication) Right() []*ConcreteInfixMultiplicationRight {
	return concrete.Right_
}

func (*ConcreteInfixMultiplication) concreteInfixOperand() {}

type ConcreteInfixMultiplicationRight struct {
	OperatorOne *ConcreteInfixMultiplicationOperator `parser:"(((IndentToken | OutdentToken)* @@ (IndentToken | OutdentToken | NewlineToken)*)"`
	OperatorTwo *ConcreteInfixMultiplicationOperator `parser:" | ((IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken)*))"`
	Operand_    *ConcretePrefixOperation             `parser:"@@"`
}

func (concrete *ConcreteInfixMultiplicationRight) Operator() *Identifier {
	if concrete.OperatorOne == nil {
		return concrete.OperatorTwo.AbstractIdentifier()
	}

	return concrete.OperatorOne.AbstractIdentifier()
}

func (concrete *ConcreteInfixMultiplicationRight) Operand() *ConcretePrefixOperation {
	return concrete.Operand_
}

type ConcreteInfixMultiplicationOperator struct {
	Identifier string `parser:"@('*':OperatorToken | '/':OperatorToken | '%':OperatorToken)"`
	Tokens     []lexer.Token
}

func (concrete *ConcreteInfixMultiplicationOperator) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Identifier,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

type ConcretePrefixOperation struct {
	Operators []*ConcreteOperator `parser:"(@@ (IndentToken | OutdentToken)*)*"`
	Operand   *ConcreteIf         `parser:"@@"`
}

func (concrete *ConcretePrefixOperation) Abstract() Expression {
	result := concrete.Operand.Abstract()

	for i := len(concrete.Operators) - 1; i >= 0; i-- {
		abstractOperator := concrete.Operators[i].AbstractIdentifier()

		result = &Call{
			Function: &Select{
				Value: result,
				Field: abstractOperator,
				Type:  parser_types.PrefixSelect,
			},

			Arguments: []Expression{},
			position: &errors.Position{
				Filename: abstractOperator.Position().Filename,
				Start:    abstractOperator.Position().Start,
				End:      result.Position().End,
			},
		}
	}

	return result
}

func (*ConcretePrefixOperation) concreteInfixOperand() {}

type ConcreteIf struct {
	Condition         ConcreteExpression         `parser:"('if':IfKeywordToken (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)*"`
	Body              *ConcreteBlock             `parser:" @@"`
	ElseIf            []*ConcreteElseIf          `parser:" (NewlineToken+ @@)*"`
	Else              *ConcreteElse              `parser:" (NewlineToken+ @@)?)"`
	AnonymousFunction *ConcreteAnonymousFunction `parser:"| @@"`
	Tokens            []lexer.Token
}

func (concrete *ConcreteIf) Abstract() Expression {
	if concrete.AnonymousFunction != nil {
		return concrete.AnonymousFunction.Abstract()
	}

	abstractFunctionFromBody := func(concrete *ConcreteBlock) *Function {
		var abstractBody *ExpressionList

		if concrete == nil {
			abstractBody = &ExpressionList{
				Children_: []Expression{
					&Identifier{
						Value:    "unit",
						position: nil,
					},
				},
			}
		} else {
			abstractBody = concrete.AbstractExpressionList()
		}

		return &Function{
			Name:       nil,
			Parameters: []*Identifier{},
			Body:       abstractBody,
		}
	}

	abstractIfFromIfOrElseIf := func(
		condition ConcreteExpression,
		thenBody *ConcreteBlock,
		elseFunction *Function,
		elsePosition *errors.Position,
		tokens []lexer.Token,
	) *Call {
		position := tokenListSyntaxTreePosition(tokens)

		if elsePosition != nil {
			position.End = elsePosition.End
		}

		return &Call{
			Function: &Identifier{
				Value:    "__if_else__",
				position: nil,
			},

			Arguments: []Expression{
				condition.Abstract(),
				abstractFunctionFromBody(thenBody),
				elseFunction,
			},

			position: position,
		}
	}

	elseBody := (*ConcreteBlock)(nil)

	if concrete.Else != nil {
		elseBody = concrete.Else.Body
	}

	current := abstractFunctionFromBody(elseBody)
	currentPosition := (*errors.Position)(nil)

	if concrete.Else != nil {
		currentPosition = tokenListSyntaxTreePosition(concrete.Else.Tokens)
	}

	for i := len(concrete.ElseIf) - 1; i >= 0; i-- {
		nextIf := abstractIfFromIfOrElseIf(
			concrete.ElseIf[i].Condition,
			concrete.ElseIf[i].Body,
			current,
			currentPosition,
			concrete.ElseIf[i].Tokens,
		)

		current = &Function{
			Name:       nil,
			Parameters: []*Identifier{},
			Body: &ExpressionList{
				Children_: []Expression{nextIf},
			},
		}

		currentPosition = nextIf.Position()
	}

	return abstractIfFromIfOrElseIf(
		concrete.Condition,
		concrete.Body,
		current,
		currentPosition,
		concrete.Tokens,
	)
}

type ConcreteElseIf struct {
	Condition ConcreteExpression `parser:"'else':ElseKeywordToken (IndentToken | OutdentToken | NewlineToken)* 'if':IfKeywordToken (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)*"`
	Body      *ConcreteBlock     `parser:"@@"`
	Tokens    []lexer.Token
}

type ConcreteElse struct {
	Body   *ConcreteBlock `parser:"'else':ElseKeywordToken (IndentToken | OutdentToken | NewlineToken)* @@"`
	Tokens []lexer.Token
}

type ConcreteAnonymousFunction struct {
	ParametersAndBody *ConcreteFunctionParametersAndBody `parser:"  (@@"`
	Call              *ConcreteCall                      `parser:" | @@)"`
	Tokens            []lexer.Token
}

func (concrete *ConcreteAnonymousFunction) Abstract() Expression {
	if concrete.Call != nil {
		return concrete.Call.Abstract()
	}

	abstractParameters, abstractBody := concrete.ParametersAndBody.Abstract()

	return &Function{
		Name:       nil,
		Parameters: abstractParameters,
		Body:       abstractBody,
		position:   tokenListSyntaxTreePosition(concrete.Tokens),
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

			result = &Call{
				Function:  result,
				Arguments: arguments,
				position: &errors.Position{
					Filename: result.Position().Filename,
					Start:    result.Position().Start,
					End: tokenSyntaxTreePosition(
						&rightHandSide.Tokens[len(rightHandSide.Tokens)-1],
					).End,
				},
			}
		} else {
			result = &Select{
				Value: result,
				Field: rightHandSide.Select.Field.AbstractIdentifier(),
				Type:  parser_types.NormalSelect,
			}
		}
	}

	return result
}

type ConcreteCallRight struct {
	Arguments *ConcreteCallArguments `parser:"(IndentToken | OutdentToken)* '(':LeftParenthesisToken (IndentToken | OutdentToken | NewlineToken)* @@? (IndentToken | OutdentToken | NewlineToken)* ')':RightParenthesisToken"`
	Select    *ConcreteSelectRight   `parser:"| @@"`
	Tokens    []lexer.Token
}

type ConcreteCallArguments struct {
	Head ConcreteExpression     `parser:"@@"`
	Tail *ConcreteCallArguments `parser:" ((IndentToken | OutdentToken | NewlineToken)* ',':CommaToken (IndentToken | OutdentToken | NewlineToken)* @@)?"`
}

type ConcreteSelect struct {
	Left  ConcretePrimary        `parser:"@@"`
	Right []*ConcreteSelectRight `parser:"@@*"`
}

func (concrete *ConcreteSelect) Abstract() Expression {
	result := concrete.Left.Abstract()

	for _, rightHandSide := range concrete.Right {
		result = &Select{
			Value: result,
			Field: rightHandSide.Field.AbstractIdentifier(),
			Type:  parser_types.NormalSelect,
		}
	}

	return result
}

type ConcreteSelectRight struct {
	Field *ConcreteIdentifier `parser:"(IndentToken | OutdentToken | NewlineToken)* '.':SelectOperatorToken (IndentToken | OutdentToken | NewlineToken)* @@"`
}

// Single-token expressions and primaries

type ConcreteParenthesized struct {
	Value ConcreteExpression `parser:"'(':LeftParenthesisToken (IndentToken | OutdentToken | NewlineToken)* @@ (IndentToken | OutdentToken | NewlineToken)* ')':RightParenthesisToken"`
}

func (concrete *ConcreteParenthesized) Abstract() Expression {
	return concrete.Value.Abstract()
}

func (*ConcreteParenthesized) primary() {}

type ConcreteTuple struct {
	Elements []ConcreteExpression `parser:"'(':LeftParenthesisToken (IndentToken | OutdentToken | NewlineToken)* (@@ ((IndentToken | OutdentToken | NewlineToken)* ',':CommaToken (IndentToken | OutdentToken | NewlineToken)* @@)+ | @@? (IndentToken | OutdentToken | NewlineToken)* ',':CommaToken) (IndentToken | OutdentToken | NewlineToken)* ')':RightParenthesisToken"`
	Tokens   []lexer.Token
}

func (concrete *ConcreteTuple) Abstract() Expression {
	abstractElements := make([]Expression, 0, len(concrete.Elements))

	for _, element := range concrete.Elements {
		abstractElements = append(abstractElements, element.Abstract())
	}

	return AbstractTuple(abstractElements, tokenListSyntaxTreePosition(concrete.Tokens))
}

func (*ConcreteTuple) primary() {}

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
	Value  string `parser:"@IdentifierToken | @OperatorToken"`
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

type ConcreteOperator struct {
	Value  string `parser:"@OperatorToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteOperator) AbstractIdentifier() *Identifier {
	return &Identifier{
		Value:    concrete.Value,
		position: tokenSyntaxTreePosition(&concrete.Tokens[0]),
	}
}

type ConcreteString struct {
	Value  string `parser:"@StringToken"`
	Tokens []lexer.Token
}

func (concrete *ConcreteString) Abstract() Expression {
	return concrete.AbstractString()
}

func (concrete *ConcreteString) AbstractString() *String {
	/*
	 * A string's token value length isn't equal to that token's length, since we've configured
	 * Participle to automatically remove the quotes.
	 */
	unadjustedPosition := tokenSyntaxTreePosition(&concrete.Tokens[0])

	return &String{
		Value: concrete.Value,
		position: &errors.Position{
			Filename: unadjustedPosition.Filename,
			Start:    unadjustedPosition.Start,
			End:      unadjustedPosition.End + 2,
		},
	}
}

func (*ConcreteString) primary() {}

var parser = participle.MustBuild[ConcreteStatementList](
	participle.Lexer(&LexerDefinition{}),
	participle.Union[ConcreteStatement](
		&ConcreteAssignment{},
		&ConcreteFunction{},
		&ConcreteInfixMiscellaneous{},
		&ConcreteStruct{},
	),

	participle.Union[ConcreteExpression](&ConcreteInfixMiscellaneous{}),
	participle.Union[ConcretePrimary](
		&ConcreteParenthesized{},
		&ConcreteTuple{},
		&ConcreteFloat{},
		&ConcreteIdentifier{},
		&ConcreteInteger{},
		&ConcreteString{},
	),

	participle.Unquote("StringToken"),
	participle.UseLookahead(participle.MaxLookahead),
)

func ParseString(path string, source string) (*ConcreteStatementList, error) {
	return parser.ParseString(path, source)
}

func tokenListSyntaxTreePosition(tokens []lexer.Token) *errors.Position {
	lastToken := tokens[len(tokens)-1]

	return &errors.Position{
		Filename: tokens[0].Pos.Filename,
		Start:    tokens[0].Pos.Offset,
		End:      lastToken.Pos.Offset + len(lastToken.Value),
	}
}

func tokenSyntaxTreePosition(token *lexer.Token) *errors.Position {
	return &errors.Position{
		Filename: token.Pos.Filename,
		Start:    token.Pos.Offset,
		End:      token.Pos.Offset + len(token.Value),
	}
}
