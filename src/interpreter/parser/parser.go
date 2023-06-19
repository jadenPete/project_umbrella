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

func newAddition(leftHandSide Expression, rightHandSide Expression) *CallExpression {
	return &CallExpression{
		Function: &SelectExpression{
			Value: leftHandSide,
			Field: &IdentifierExpression{"__plus__"},
		},
		Arguments: []Expression{rightHandSide},
	}
}

func newBalancedAdditionFromSummands(summands []Expression) *CallExpression {
	if len(summands) == 2 {
		return newAddition(summands[0], summands[1])
	}

	if len(summands) == 3 {
		return newAddition(newAddition(summands[0], summands[1]), summands[2])
	}

	middle := len(summands) / 2

	return newAddition(
		newBalancedAdditionFromSummands(summands[:middle]),
		newBalancedAdditionFromSummands(summands[middle:]),
	)
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
	Value Expression
	Field *IdentifierExpression
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
	additionOperatorTokenCode MatcherCode = iota + 1
	assignmentOperatorTokenCode
	commaTokenCode
	floatTokenCode
	identifierTokenCode
	integerTokenCode
	leftParenthesisTokenCode
	rightParenthesisTokenCode
	newlineTokenCode
	selectOperatorTokenCode
	stringTokenCode

	additionExpressionCode
	assignmentExpressionCode
	expressionListExpressionCode
	callExpressionCode
	floatExpressionCode
	identifierExpressionCode
	integerExpressionCode
	selectExpressionCode
	stringExpressionCode
)

var tokenTypeCodes = map[TokenType]MatcherCode{
	AdditionOperatorToken:   additionOperatorTokenCode,
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
	additionExpressionCode,
	assignmentExpressionCode,
	callExpressionCode,
	floatExpressionCode,
	identifierExpressionCode,
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
	patterns: []*ExhaustiveMatchPattern{
		{
			type_: selectExpressionCode,
			matcher: CompileMatcher(
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
			type_: callExpressionCode,
			matcher: CompileMatcher(
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
			type_: additionExpressionCode,
			matcher: CompileMatcher(
				fmt.Sprintf(
					`(?:%s{0}?{1}{0}?)+%s`,
					standaloneExpressionRegex(2),
					standaloneExpressionRegex(2),
				),
				append(
					[]MatcherCode{
						newlineTokenCode,
						additionOperatorTokenCode,
					},

					standaloneExpressionCodes...,
				)...,
			),
		},

		{
			type_: assignmentExpressionCode,
			matcher: CompileMatcher(
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
			type_:   floatExpressionCode,
			matcher: CompileMatcher(`{0}`, floatTokenCode),
		},

		{
			type_:   identifierExpressionCode,
			matcher: CompileMatcher(`{0}`, identifierTokenCode),
		},

		{
			type_:   integerExpressionCode,
			matcher: CompileMatcher(`{0}`, integerTokenCode),
		},

		{
			type_:   stringExpressionCode,
			matcher: CompileMatcher(`{0}`, stringTokenCode),
		},

		{
			type_: expressionListExpressionCode,
			matcher: CompileMatcher(
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
		input = append(input, tokenTypeCodes[token.type_])
	}

	tree := parserExhaustiveMatcher.MatchTree(input)

	if tree == nil {
		return nil
	}

	return parser.parseMatchTree(tree)
}

func (parser *Parser) parseMatchTree(tree *common.Tree[*ExhaustiveMatch]) Expression {
	switch tree.Value.type_ {
	case additionExpressionCode:
		/*
		 * Addition is assumed to be associative.
		 * We take advantage of this to parallelize it as much as possible.
		 */
		return newBalancedAdditionFromSummands(
			parseParsableMatchTrees[Expression](parser, tree.Children),
		)

	case assignmentExpressionCode:
		i := tree.Value.subgroups[0][0]

		names := parseParsableMatchTrees[*IdentifierExpression](parser, tree.Children[:i])

		return &AssignmentExpression{
			Names: names,
			Value: parser.parseMatchTree(tree.Children[tree.Value.subgroups[0][0]]),
		}

	case callExpressionCode:
		argumentSubgroup := tree.Value.subgroups[0]

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

		if len(tree.Children) > 0 && tree.Children[0].Value.type_ == newlineTokenCode {
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
		token := parser.Tokens[tree.Value.start]

		value, _ := strconv.ParseFloat(parser.FileContent[token.start:token.end], 32)

		return &FloatExpression{
			Value: value,
		}

	case identifierExpressionCode:
		token := parser.Tokens[tree.Value.start]

		return &IdentifierExpression{
			Content: parser.FileContent[token.start:token.end],
		}

	case integerExpressionCode:
		token := parser.Tokens[tree.Value.start]

		value, _ := strconv.ParseInt(parser.FileContent[token.start:token.end], 10, 64)

		return &IntegerExpression{
			Value: value,
		}

	case selectExpressionCode:
		result := parser.parseMatchTree(tree.Children[0])

		identifiers := parseParsableMatchTrees[*IdentifierExpression](parser, tree.Children[1:])

		for _, identifier := range identifiers {
			result = &SelectExpression{
				Value: result,
				Field: identifier,
			}
		}

		return result

	case stringExpressionCode:
		token := parser.Tokens[tree.Value.start]

		return &StringExpression{
			Content: parser.FileContent[token.start+1 : token.end-1],
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
