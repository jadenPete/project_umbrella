package parser

import (
	"fmt"
	"project_umbrella/interpreter/common"
)

type Expression interface {
	Visit(ExpressionVisitor)
}

type ExpressionVisitor struct {
	VisitAssignment               func(*AssignmentExpression)
	VisitExpressionListExpression func(*ExpressionListExpression)
	VisitCall                     func(*CallExpression)
	VisitIdentifier               func(*IdentifierExpression)
	VisitString                   func(*StringExpression)
}

type AssignmentExpression struct {
	Name  *IdentifierExpression
	Value Expression
}

func (assignment *AssignmentExpression) Visit(visitor ExpressionVisitor) {
	visitor.VisitAssignment(assignment)
}

type ExpressionListExpression struct {
	Children []Expression
}

func (expressionList *ExpressionListExpression) Visit(visitor ExpressionVisitor) {
	visitor.VisitExpressionListExpression(expressionList)
}

type CallExpression struct {
	Identifier *IdentifierExpression
	Argument   Expression
}

func (call *CallExpression) Visit(visitor ExpressionVisitor) {
	visitor.VisitCall(call)
}

type IdentifierExpression struct {
	Content string
}

func (identifier *IdentifierExpression) Visit(visitor ExpressionVisitor) {
	visitor.VisitIdentifier(identifier)
}

type StringExpression struct {
	Content string
}

func (string_ *StringExpression) Visit(visitor ExpressionVisitor) {
	visitor.VisitString(string_)
}

const (
	assignmentOperatorTokenCode MatcherCode = iota + 1
	identifierTokenCode
	leftParenthesisTokenCode
	rightParenthesisTokenCode
	newlineTokenCode
	stringTokenCode

	assignmentExpressionCode
	expressionListExpressionCode
	callExpressionCode
	identifierExpressionCode
	stringExpressionCode
)

func tokenCode(token *Token) MatcherCode {
	switch token.type_ {
	case StringToken:
		return stringTokenCode
	case AssignmentOperatorToken:
		return assignmentOperatorTokenCode
	case IdentifierToken:
		return identifierTokenCode
	case LeftParenthesisToken:
		return leftParenthesisTokenCode
	case RightParenthesisToken:
		return rightParenthesisTokenCode
	case NewlineToken:
		return newlineTokenCode
	}

	return 0
}

func standaloneExpressionRegex(startingSubstitutionIndex int) string {
	return fmt.Sprintf(
		`[{%d}{%d}{%d}{%d}]`,
		startingSubstitutionIndex,
		startingSubstitutionIndex+1,
		startingSubstitutionIndex+2,
		startingSubstitutionIndex+3,
	)
}

var standaloneExpressionCodes = []MatcherCode{
	assignmentExpressionCode,
	callExpressionCode,
	identifierExpressionCode,
	stringExpressionCode,
}

var parserExhaustiveMatcher ExhaustiveMatcher = ExhaustiveMatcher{
	patterns: []*ExhaustiveMatchPattern{
		{
			type_: assignmentExpressionCode,
			matcher: CompileMatcher(
				fmt.Sprintf(`{0}{1}*{2}{1}*(%s)`, standaloneExpressionRegex(3)),
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
			type_: callExpressionCode,
			matcher: CompileMatcher(
				fmt.Sprintf(`{0}{1}{2}*(%s){2}*{3}`, standaloneExpressionRegex(4)),
				append(
					[]MatcherCode{
						identifierExpressionCode,
						leftParenthesisTokenCode,
						newlineTokenCode,
						rightParenthesisTokenCode,
					},

					standaloneExpressionCodes...,
				)...,
			),
		},

		{
			type_: expressionListExpressionCode,
			matcher: CompileMatcher(
				fmt.Sprintf(`^(?:%s{0}+)*$`, standaloneExpressionRegex(1)),
				append([]MatcherCode{newlineTokenCode}, standaloneExpressionCodes...)...,
			),
		},

		{
			type_:   identifierExpressionCode,
			matcher: CompileMatcher(`{0}`, identifierTokenCode),
		},

		{
			type_:   stringExpressionCode,
			matcher: CompileMatcher(`{0}`, stringTokenCode),
		},
	},
}

type Parser struct {
	fileContent string
	tokens      []*Token
}

func NewParser(fileContent string, tokens []*Token) *Parser {
	/*
	 * The caller of Parser.Parse should expect a ExpressionListExpression, which requires
	 * the newline to be used as a delimeter, not a separator.
	 */
	if len(fileContent) > 0 && fileContent[len(fileContent)-1] != '\n' {
		fileContent = fileContent + " "

		newTokens := make([]*Token, len(tokens), len(tokens)+1)

		copy(newTokens, tokens)

		newTokens = append(newTokens, &Token{
			type_: NewlineToken,
			start: len(fileContent) - 1,
			end:   len(fileContent),
		})

		tokens = newTokens
	}

	return &Parser{fileContent, tokens}
}

func (parser *Parser) Parse() Expression {
	input := make([]MatcherCode, 0, len(parser.tokens))

	for _, token := range parser.tokens {
		input = append(input, tokenCode(token))
	}

	tree := parserExhaustiveMatcher.MatchTree(input)

	if tree == nil {
		return nil
	}

	return parser.parseMatchTree(tree)
}

func (parser *Parser) parseMatchTree(tree *common.Tree[*ExhaustiveMatch]) Expression {
	switch tree.Value.type_ {
	case assignmentExpressionCode:
		return &AssignmentExpression{
			Name:  parser.parseMatchTree(tree.Children[0]).(*IdentifierExpression),
			Value: parser.parseMatchTree(tree.Children[tree.Value.subgroups[0][0]]),
		}

	case expressionListExpressionCode:
		children := make([]Expression, 0, len(tree.Children)/2)

		for i := 0; i < len(tree.Children); i += 2 {
			children = append(children, parser.parseMatchTree(tree.Children[i]))
		}

		return &ExpressionListExpression{children}

	case callExpressionCode:
		return &CallExpression{
			Identifier: parser.parseMatchTree(tree.Children[0]).(*IdentifierExpression),
			Argument:   parser.parseMatchTree(tree.Children[tree.Value.subgroups[0][0]]),
		}

	case identifierExpressionCode:
		token := parser.tokens[tree.Value.start]

		return &IdentifierExpression{
			Content: parser.fileContent[token.start:token.end],
		}

	case stringExpressionCode:
		token := parser.tokens[tree.Value.start]

		return &StringExpression{
			Content: parser.fileContent[token.start+1 : token.end-1],
		}
	}

	return nil
}
