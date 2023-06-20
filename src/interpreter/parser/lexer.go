package parser

import (
	"fmt"
	"strings"

	"project_umbrella/interpreter/common"
)

type TokenType MatcherCode
type Token struct {
	Type  TokenType
	Start int
	End   int
}

const (
	StringToken TokenType = iota + 1
	AssignmentOperatorToken
	CommaToken
	FloatToken
	IdentifierToken
	IndentToken
	OutdentToken
	NewlineToken
	SpaceToken
	IntegerToken
	LeftParenthesisToken
	RightParenthesisToken
	SelectOperatorToken
)

var matcher = ExhaustiveMatcher{
	[]*ExhaustiveMatchPattern{
		{
			MatcherCode(StringToken),
			CompileMatcher(`"[^"]*"`),
		},

		{
			MatcherCode(AssignmentOperatorToken),
			CompileMatcher(`=`),
		},

		{
			MatcherCode(CommaToken),
			CompileMatcher(`,`),
		},

		{
			MatcherCode(FloatToken),
			CompileMatcher(`(?:\+|-)?(?:\d+\.\d*|\.\d+)`),
		},

		{
			MatcherCode(LeftParenthesisToken),
			CompileMatcher(`\(`),
		},

		{
			MatcherCode(RightParenthesisToken),
			CompileMatcher(`\)`),
		},

		{
			MatcherCode(NewlineToken),
			CompileMatcher(`\n`),
		},

		{
			MatcherCode(SelectOperatorToken),
			CompileMatcher(`\.`),
		},

		{
			MatcherCode(SpaceToken),
			CompileMatcher(`[\t ]+`),
		},

		{
			MatcherCode(IdentifierToken),
			CompileMatcher(`[^\t\n ="(),.]*[^\t\n ="(),.\d]+[^\t\n ="(),.]*`),
		},

		{
			MatcherCode(IntegerToken),
			CompileMatcher(`(?:\+|-)?\d+`),
		},
	},
}

type Lexer struct {
	FileContent string
}

func isLineBlank(line string) bool {
	for _, character := range line {
		if character != '\t' && character != ' ' {
			return false
		}
	}

	return len(line) > 0
}

/*
 * Tokenize the input file content, converting it into a slice of tokens.
 *
 * If tokenization failed, this function returns `nil`.
 */
func (lexer *Lexer) Parse() []*Token {
	if len(lexer.FileContent) == 0 {
		return make([]*Token, 0)
	}

	matches := matcher.MatchWithInitial(MatcherInput(lexer.FileContent), lexer.parseIndentation())

	if matches == nil {
		return nil
	}

	result := make([]*Token, 0)

	for _, match := range matches {
		if match.Type != MatcherCode(SpaceToken) {
			result = append(result, &Token{
				Type:  TokenType(match.Type),
				Start: match.Start,
				End:   match.End,
			})
		}
	}

	return result
}

func (lexer *Lexer) parseIndentation() []*ExhaustiveMatch {
	result := make([]*ExhaustiveMatch, 0)

	var indentCharacter rune

	indentLength := 0
	lastIndentCount := 0

	fileLines := strings.Split(lexer.FileContent, "\n")
	fileOffset := 0

	for i, line := range fileLines {
		lineCharacters := []rune(line)

		if !isLineBlank(line) {
			if indentCharacter == 0 &&
				len(lineCharacters) > 0 &&
				(lineCharacters[0] == '\t' || lineCharacters[0] == ' ') {
				indentCharacter = lineCharacters[0]
			}

			if indentCharacter != 0 {
				indentCount := 0

				for indentCount < len(lineCharacters) &&
					lineCharacters[indentCount] == indentCharacter {
					indentCount++
				}

				if indentLength == 0 {
					indentLength = indentCount
				} else if indentCount%indentLength > 0 {
					var indentCharacterName string

					if indentCharacter == '\t' {
						indentCharacterName = "tabs"
					} else {
						indentCharacterName = "spaces"
					}

					panic(
						fmt.Sprintf(
							"Inconsistent indentation; you indent using %[2]d %[1]s, but this line is prefixed with %[3]d %[1]s",
							indentCharacterName,
							indentLength,
							indentCount,
						),
					)
				}

				indentCount /= indentLength

				indentDelta := common.Abs(indentCount - lastIndentCount)

				var indentMatcherCode MatcherCode

				if indentCount >= lastIndentCount {
					indentMatcherCode = MatcherCode(IndentToken)
				} else {
					indentMatcherCode = MatcherCode(OutdentToken)
				}

				for i := 0; i < indentDelta; i++ {
					result = append(result, &ExhaustiveMatch{
						Type:      indentMatcherCode,
						Start:     fileOffset + indentCount*indentLength,
						End:       fileOffset + indentCount*indentLength,
						Subgroups: make([][2]int, 0),
					})
				}

				lastIndentCount = indentCount
			}
		}

		start := fileOffset + lastIndentCount*indentLength
		end := fileOffset + len(line)

		if i < len(fileLines)-1 {
			end++
		}

		if end-start > 0 {
			result = append(result, &ExhaustiveMatch{
				Type:      UnrecognizedMatcherCode,
				Start:     start,
				End:       end,
				Subgroups: make([][2]int, 0),
			})
		}

		fileOffset += len(line) + 1
	}

	return result
}
