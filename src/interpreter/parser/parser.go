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
	indentTokenCode
	outdentTokenCode
	newlineTokenCode
	integerTokenCode
	leftParenthesisTokenCode
	rightParenthesisTokenCode
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
	IndentToken:             indentTokenCode,
	OutdentToken:            outdentTokenCode,
	NewlineToken:            newlineTokenCode,
	IntegerToken:            integerTokenCode,
	LeftParenthesisToken:    leftParenthesisTokenCode,
	RightParenthesisToken:   rightParenthesisTokenCode,
	SelectOperatorToken:     selectOperatorTokenCode,
	StringToken:             stringTokenCode,
}

var formattingExpressionCodes = []MatcherCode{indentTokenCode, outdentTokenCode, newlineTokenCode}
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

func compositeExpressionRegex(expressionCodes []MatcherCode, startingSubstitutionIndex int) string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(`[`)

	for i := startingSubstitutionIndex; i < startingSubstitutionIndex+len(expressionCodes); i++ {
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
				fmt.Sprintf(
					`%[1]s(?:%[2]s*{0}%[2]s*{1})+`,
					compositeExpressionRegex(standaloneExpressionCodes, 2),
					compositeExpressionRegex(
						formattingExpressionCodes,
						len(standaloneExpressionCodes)+2,
					),
				),

				append(
					append(
						[]MatcherCode{selectOperatorTokenCode, identifierExpressionCode},
						standaloneExpressionCodes...,
					),

					formattingExpressionCodes...,
				)...,
			),
		},

		{
			Type: callExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`%[1]s[{0}{1}]?{2}%[2]s*(%[1]s(?:%[2]s*{3}%[1]s)*)?%[2]s*{4}`,
					compositeExpressionRegex(standaloneExpressionCodes, 5),
					compositeExpressionRegex(
						formattingExpressionCodes,
						len(standaloneExpressionCodes)+5,
					),
				),

				append(
					append(
						[]MatcherCode{
							indentTokenCode,
							outdentTokenCode,
							leftParenthesisTokenCode,
							commaTokenCode,
							rightParenthesisTokenCode,
						},

						standaloneExpressionCodes...,
					),

					formattingExpressionCodes...,
				)...,
			),
		},

		{
			Type: infixCallExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`%[1]s(?:%[2]s*{0}%[2]s*%[1]s)+`,
					compositeExpressionRegex(standaloneExpressionCodes, 1),
					compositeExpressionRegex(
						formattingExpressionCodes,
						len(standaloneExpressionCodes)+1,
					),
				),

				append(
					append([]MatcherCode{identifierExpressionCode}, standaloneExpressionCodes...),
					formattingExpressionCodes...,
				)...,
			),
		},

		{
			Type: assignmentExpressionCode,
			Matcher: CompileMatcher(
				fmt.Sprintf(
					`(?:{0}%[1]s*{1}%[1]s*)+(%[2]s)`,
					compositeExpressionRegex(formattingExpressionCodes, 2),
					compositeExpressionRegex(
						standaloneExpressionCodes,
						len(formattingExpressionCodes)+2,
					),
				),

				append(
					append(
						[]MatcherCode{identifierExpressionCode, assignmentOperatorTokenCode},
						formattingExpressionCodes...,
					),

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
					`^{0}*(?:(?:%[1]s{0}+)*%[1]s{0}*)?$`,
					compositeExpressionRegex(standaloneExpressionCodes, 1),
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

/*
 * This function was added to solve the problem of what I've deemed hanging expressions.
 * Consider the following program.
 *
 * ```
 * message =
 *     "Hello, world!"
 *
 * println(message)
 * ```
 *
 * Although the assignment to `message` and call to `println` will be parsed correctly,
 * the former is implicitly followed by a newline, then outdent token. This is a problem because
 * conjoining both expressions into a final expression list requires they be separated only by
 * newlines.
 *
 * To solve this issue, we "float" indentation parsed in an expression out and to the right of that
 * expression, cancelling it out with converse indentation where possible. In the above example, the
 * indent token inside the assignment would float outward and cancel with the outdent following the
 * assignment, allowing for the formation of a proper expression list.
 */
func addIndentationAfterHangingExpressions(
	unmatched []*common.Tree[*ExhaustiveMatch],
	newIndices []int,
) []*common.Tree[*ExhaustiveMatch] {
	result := make([]*common.Tree[*ExhaustiveMatch], 0)

	indentsNeeded := 0
	i := 0

	for j, tree := range unmatched {
		if tree.Value.Type == indentTokenCode {
			indentsNeeded++
		} else if tree.Value.Type == outdentTokenCode {
			indentsNeeded--
		} else {
			if tree.Value.Type != newlineTokenCode {
				var indentMatcherCode MatcherCode

				if indentsNeeded >= 0 {
					indentMatcherCode = indentTokenCode
				} else {
					indentMatcherCode = outdentTokenCode
				}

				for i := 0; i < common.Abs(indentsNeeded); i++ {
					result = append(result, &common.Tree[*ExhaustiveMatch]{
						Children: make([]*common.Tree[*ExhaustiveMatch], 0),
						Value: &ExhaustiveMatch{
							Type:      indentMatcherCode,
							Start:     tree.Value.Start,
							End:       tree.Value.Start,
							Subgroups: make([][2]int, 0),
						},
					})
				}

				indentsNeeded = 0

				if i < len(newIndices) && newIndices[i] == j {
					for _, child := range tree.Children {
						if child.Value.Type == indentTokenCode {
							indentsNeeded++
						} else if child.Value.Type == outdentTokenCode {
							indentsNeeded--
						}
					}

					i++
				}
			}

			result = append(result, tree)
		}
	}

	return result
}

func (parser *Parser) Parse() Expression {
	input := make([]MatcherCode, 0, len(parser.Tokens))

	for _, token := range parser.Tokens {
		input = append(input, tokenTypeCodes[token.Type])
	}

	tree := parserExhaustiveMatcher.MatchTreeWithTransformation(
		input,
		addIndentationAfterHangingExpressions,
	)

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
		return &ExpressionListExpression{parseParsableMatchTrees[Expression](parser, tree.Children)}

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

func parseParsableMatchTrees[T Expression](
	parser *Parser,
	matchTrees []*common.Tree[*ExhaustiveMatch],
) []T {
	parsed := make([]T, 0)

	for _, matchTree := range matchTrees {
		if expression := parser.parseMatchTree(matchTree); expression != nil {
			parsed = append(parsed, expression.(T))
		}
	}

	return parsed
}
