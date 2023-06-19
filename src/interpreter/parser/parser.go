package parser

import (
	"fmt"
	"strconv"
	"strings"

	"project_umbrella/interpreter/common"
)

type Expression interface {
	Visit(*ExpressionVisitor)
}

type ExpressionVisitor struct {
	VisitAssignment               func(*AssignmentExpression)
	VisitExpressionListExpression func(*ExpressionListExpression)
	VisitCall                     func(*CallExpression)
	VisitFloat                    func(*FloatExpression)
	VisitIdentifier               func(*IdentifierExpression)
	VisitInteger                  func(*IntegerExpression)
	VisitSelect                   func(*SelectExpression)
	VisitString                   func(*StringExpression)
}

type AssignmentExpression struct {
	Names []*IdentifierExpression
	Value Expression
}

func (assignment *AssignmentExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitAssignment(assignment)
}

type ExpressionListExpression struct {
	Children []Expression
}

func (expressionList *ExpressionListExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitExpressionListExpression(expressionList)
}

type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

/*
 * Apply the shunting yard algorithm on a parsed operation
 * (infix expression comprising operands and identifier operators), reducing it to an AST
 * containing `CallExpression`s and `SelectExpression`s.
 *
 * Operator precedence is determined according to operators' first character.
 * Those at the top of the following list are of higher precedence.
 * Characters on the same row are of equal precedence.
 *
 * 1. *, /
 * 2. +, -
 * 3. Every other character
 */
func newChainedInfixCallExpression(operation []Expression) *CallExpression {
	operandStack := make([]Expression, 0, len(operation)/2+1)
	operatorStack := make([]*IdentifierExpression, 0, len(operation)/2)

	evaluateTopOperation := func() {
		topOperation := &CallExpression{
			Function: &SelectExpression{
				Value:   operandStack[len(operandStack)-2],
				Field:   operatorStack[len(operatorStack)-1],
				IsInfix: true,
			},

			Arguments: []Expression{operandStack[len(operandStack)-1]},
		}

		operandStack = operandStack[:len(operandStack)-2]
		operatorStack = operatorStack[:len(operatorStack)-1]
		operandStack = append(operandStack, topOperation)
	}

	for i, expression := range operation {
		if i%2 == 0 {
			operandStack = append(operandStack, expression)
		} else {
			operator := expression.(*IdentifierExpression)

			for len(operatorStack) > 0 &&
				operatorPrecedence(operatorStack[len(operatorStack)-1]) <=
					operatorPrecedence(operator) {
				evaluateTopOperation()
			}

			operatorStack = append(operatorStack, operator)
		}
	}

	for len(operatorStack) > 0 {
		evaluateTopOperation()
	}

	return operandStack[0].(*CallExpression)
}

func operatorPrecedence(operator *IdentifierExpression) int {
	switch []rune(operator.Content)[0] {
	case '*', '/':
		return 0

	case '+', '-':
		return 1

	default:
		return 2
	}
}

func (call *CallExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitCall(call)
}

type FloatExpression struct {
	Value float64
}

func (float *FloatExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitFloat(float)
}

type IdentifierExpression struct {
	Content string
}

func (identifier *IdentifierExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitIdentifier(identifier)
}

type IntegerExpression struct {
	Value int64
}

func (integer *IntegerExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitInteger(integer)
}

type SelectExpression struct {
	Value   Expression
	Field   *IdentifierExpression
	IsInfix bool
}

func (select_ *SelectExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitSelect(select_)
}

type StringExpression struct {
	Content string
}

func (string_ *StringExpression) Visit(visitor *ExpressionVisitor) {
	visitor.VisitString(string_)
}

const (
	assignmentOperatorTokenCode MatcherCode = iota + 1
	commaTokenCode
	floatTokenCode
	identifierTokenCode
	integerTokenCode
	leftParenthesisTokenCode
	rightParenthesisTokenCode
	newlineTokenCode
	selectOperatorTokenCode
	stringTokenCode

	assignmentExpressionCode
	expressionListExpressionCode
	callExpressionCode
	floatExpressionCode
	identifierExpressionCode
	infixCallExpressionCode
	integerExpressionCode
	selectExpressionCode
	stringExpressionCode
)

var tokenTypeCodes = map[TokenType]MatcherCode{
	AssignmentOperatorToken: assignmentOperatorTokenCode,
	CommaToken:              commaTokenCode,
	FloatToken:              floatTokenCode,
	IdentifierToken:         identifierTokenCode,
	IntegerToken:            integerTokenCode,
	LeftParenthesisToken:    leftParenthesisTokenCode,
	RightParenthesisToken:   rightParenthesisTokenCode,
	NewlineToken:            newlineTokenCode,
	SelectOperatorToken:     selectOperatorTokenCode,
	StringToken:             stringTokenCode,
}

var standaloneExpressionCodes = []MatcherCode{
	assignmentExpressionCode,
	callExpressionCode,
	floatExpressionCode,
	identifierExpressionCode,
	infixCallExpressionCode,
	integerExpressionCode,
	selectExpressionCode,
	stringExpressionCode,
}

func standaloneExpressionRegex(startingSubstitutionIndex int) string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(`[`)

	for i := startingSubstitutionIndex; i < startingSubstitutionIndex+len(standaloneExpressionCodes); i += 1 {
		stringBuilder.WriteString(fmt.Sprintf(`{%d}`, i))
	}

	stringBuilder.WriteString(`]`)

	return stringBuilder.String()
}

var parserExhaustiveMatcher ExhaustiveMatcher = ExhaustiveMatcher{
	[]*ExhaustiveMatchPattern{
		{
			Type: selectExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(`%s(?:{0}?{1}{0}?{2})+`, standaloneExpressionRegex(3)),
				append(
					[]MatcherCode{
						newlineTokenCode,
						selectOperatorTokenCode,
						identifierExpressionCode,
					},

					standaloneExpressionCodes...,
				)...,
			),
		},

		{
			Type: callExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`%s{0}?{1}{0}?(%s(?:{0}?{2}%s)*)?{0}?{3}`,
					standaloneExpressionRegex(4),
					standaloneExpressionRegex(4),
					standaloneExpressionRegex(4),
				),
				append(
					[]MatcherCode{
						newlineTokenCode,
						leftParenthesisTokenCode,
						commaTokenCode,
						rightParenthesisTokenCode,
					},

					standaloneExpressionCodes...,
				)...,
			),
		},

		{
			Type: infixCallExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`%s(?:{0}%s)+`,
					standaloneExpressionRegex(1),
					standaloneExpressionRegex(1),
				),
				append([]MatcherCode{identifierExpressionCode}, standaloneExpressionCodes...)...,
			),
		},

		{
			Type: assignmentExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(`(?:{0}{1}?{2}{1}?)+(%s)`, standaloneExpressionRegex(3)),
				append(
					[]MatcherCode{
						identifierExpressionCode,
						newlineTokenCode,
						assignmentOperatorTokenCode,
					},

					standaloneExpressionCodes...,
				)...,
			),
		},

		{
			Type:    floatExpressionCode,
			Matcher: CompileMatcher(`{0}`, floatTokenCode),
		},

		{
			Type:    identifierExpressionCode,
			Matcher: CompileMatcher(`{0}`, identifierTokenCode),
		},

		{
			Type:    integerExpressionCode,
			Matcher: CompileMatcher(`{0}`, integerTokenCode),
		},

		{
			Type:    stringExpressionCode,
			Matcher: CompileMatcher(`{0}`, stringTokenCode),
		},

		{
			Type: expressionListExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`^{0}?(?:(?:%s{0})*%s{0}?)?$`,
					standaloneExpressionRegex(1),
					standaloneExpressionRegex(1),
				),
				append([]MatcherCode{newlineTokenCode}, standaloneExpressionCodes...)...,
			),
		},
	},
}

type Parser struct {
	FileContent string
	Tokens      []*Token
}

func (parser *Parser) Parse() Expression {
	input := make([]MatcherCode, 0, len(parser.Tokens))

	for _, token := range parser.Tokens {
		input = append(input, tokenTypeCodes[token.Type])
	}

	tree := parserExhaustiveMatcher.MatchTree(input)

	if tree == nil {
		return nil
	}

	return parser.parseMatchTree(tree)
}

func (parser *Parser) parseMatchTree(tree *common.Tree[*ExhaustiveMatch]) Expression {
	switch tree.Value.Type {
	case assignmentExpressionCode:
		i := tree.Value.Subgroups[0][0]

		names := parseParsableMatchTrees[*IdentifierExpression](parser, tree.Children[:i])

		return &AssignmentExpression{
			Names: names,
			Value: parser.parseMatchTree(tree.Children[tree.Value.Subgroups[0][0]]),
		}

	case callExpressionCode:
		argumentSubgroup := tree.Value.Subgroups[0]

		var arguments []Expression

		if argumentSubgroup[0] == -1 {
			arguments = []Expression{}
		} else {
			arguments = parseParsableMatchTrees[Expression](
				parser,
				tree.Children[argumentSubgroup[0]:argumentSubgroup[1]],
			)
		}

		return &CallExpression{
			Function:  parser.parseMatchTree(tree.Children[0]),
			Arguments: arguments,
		}

	case expressionListExpressionCode:
		children := make([]Expression, 0, (len(tree.Children)+1)/2)

		var i int

		if len(tree.Children) > 0 && tree.Children[0].Value.Type == newlineTokenCode {
			i = 1
		} else {
			i = 0
		}

		for i < len(tree.Children) {
			children = append(children, parser.parseMatchTree(tree.Children[i]))

			i += 2
		}

		return &ExpressionListExpression{children}

	case floatExpressionCode:
		token := parser.Tokens[tree.Value.Start]

		value, _ := strconv.ParseFloat(parser.FileContent[token.Start:token.End], 32)

		return &FloatExpression{
			Value: value,
		}

	case identifierExpressionCode:
		token := parser.Tokens[tree.Value.Start]

		return &IdentifierExpression{
			Content: parser.FileContent[token.Start:token.End],
		}

	case infixCallExpressionCode:
		parsedChildren := make([]Expression, 0, len(tree.Children))

		for _, child := range tree.Children {
			parsedChildren = append(parsedChildren, parser.parseMatchTree(child))
		}

		return newChainedInfixCallExpression(parsedChildren)

	case integerExpressionCode:
		token := parser.Tokens[tree.Value.Start]

		value, _ := strconv.ParseInt(parser.FileContent[token.Start:token.End], 10, 64)

		return &IntegerExpression{
			Value: value,
		}

	case selectExpressionCode:
		result := parser.parseMatchTree(tree.Children[0])

		identifiers := parseParsableMatchTrees[*IdentifierExpression](parser, tree.Children[1:])

		for _, identifier := range identifiers {
			result = &SelectExpression{
				Value:   result,
				Field:   identifier,
				IsInfix: false,
			}
		}

		return result

	case stringExpressionCode:
		token := parser.Tokens[tree.Value.Start]

		return &StringExpression{
			Content: parser.FileContent[token.Start+1 : token.End-1],
		}
	}

	return nil
}

func parseParsableMatchTrees[T Expression](parser *Parser, matcheTrees []*common.Tree[*ExhaustiveMatch]) []T {
	parsed := make([]T, 0)

	for _, matchTree := range matcheTrees {
		if expression := parser.parseMatchTree(matchTree); expression != nil {
			parsed = append(parsed, expression.(T))
		}
	}

	return parsed
}
