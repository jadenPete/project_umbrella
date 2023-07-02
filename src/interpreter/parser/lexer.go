package parser

import (
	"fmt"
	"strings"
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
	ColonToken
	CommaToken
	FloatToken
	FunctionKeywordToken
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
			MatcherCode(ColonToken),
			CompileMatcher(`:`),
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
			MatcherCode(FunctionKeywordToken),
			CompileMatcher(`^fn$`),
		},

		{
			MatcherCode(IntegerToken),
			CompileMatcher(`^(?:\+|-)?\d+$`),
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
			CompileMatcher(`[^\t\n ="(),.]+`),
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

	return true
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

/*
 * Before we tokenize the input, we replace indentation with indent and outde nt meta-tokens, which
 * effectively act as opening and closing tags. Iterating through each line, we record the
 * difference in indentation between it and the previous one, adding indents or outdents to it as
 * appropriate.
 *
 * For example, the following input...
 *
 * fn print_hello_world():
 *     print("Hello, world!")
 *
 * print_hello_world()
 *
 * ...becomes the following.
 *
 * fn print_hello_world():\n
 * ->print("Hello, world!")<-\n
 * \n
 * print_hello_world()\n
 *
 * Note that the indentation character (tab or space) and length is auto-determined and checked for
 * consistency. Additionally, added indent tokens always precede lines, while outdent tokens always
 * succeed them.
 */
func (lexer *Lexer) parseIndentation() []*ExhaustiveMatch {
	result := make([]*ExhaustiveMatch, 0)

	addMatchToResult := func(type_ MatcherCode, start int, end int) {
		result = append(result, &ExhaustiveMatch{
			Type:      type_,
			Start:     start,
			End:       end,
			Subgroups: make([][2]int, 0),
		})
	}

	endOfLastMatch := func() int {
		if len(result) == 0 {
			return 0
		}

		return result[len(result)-1].End
	}

	var indentCharacter rune

	fileOffset := 0
	indentLength := 0
	lastIndentCount := 0

	for _, line := range strings.Split(lexer.FileContent, "\n") {
		isLineBlank := isLineBlank(line)
		indentCount := 0

		if !isLineBlank {
			lineCharacters := []rune(line)

			if indentCharacter == 0 && (lineCharacters[0] == '\t' || lineCharacters[0] == ' ') {
				indentCharacter = lineCharacters[0]
			}

			if indentCharacter != 0 {
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
			}
		}

		if !isLineBlank {
			endOfLastUnrecognizedMatch := endOfLastMatch()
			lineStart := fileOffset + indentCount*indentLength

			if indentCount < lastIndentCount {
				for i := 0; i < lastIndentCount-indentCount; i++ {
					addMatchToResult(
						MatcherCode(OutdentToken),
						endOfLastUnrecognizedMatch,
						endOfLastUnrecognizedMatch,
					)
				}
			}

			if endOfLastUnrecognizedMatch > 0 {
				addMatchToResult(UnrecognizedMatcherCode, endOfLastUnrecognizedMatch, fileOffset)
			}

			if indentCount > lastIndentCount {
				for i := 0; i < indentCount-lastIndentCount; i++ {
					addMatchToResult(MatcherCode(IndentToken), lineStart, lineStart)
				}
			}

			lastIndentCount = indentCount

			addMatchToResult(UnrecognizedMatcherCode, lineStart, fileOffset+len(line))
		}

		fileOffset += len(line) + 1
	}

	if endOfLastMatch() < len(lexer.FileContent) {
		addMatchToResult(UnrecognizedMatcherCode, endOfLastMatch(), len(lexer.FileContent))
	}

	return result
}
