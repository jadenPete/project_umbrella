package parser

import (
	"io"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/lexer_errors"
)

const (
	AssignmentOperatorToken lexer.TokenType = iota + 1
	ColonToken
	CommaToken
	ElseToken
	ElseIfToken
	IfToken
	FloatToken
	FunctionKeywordToken
	IdentifierToken
	IndentToken
	OutdentToken
	IntegerToken
	LeftParenthesisToken
	RightParenthesisToken
	NewlineToken
	OperatorToken
	SelectOperatorToken
	SpaceToken
	StringToken
)

var matcher = ExhaustiveMatcher{
	[]*ExhaustiveMatchPattern{
		/*
		 * Strings are parsed first because they can contain nearly any character,
		 * and their content shouldn't be mistaken for other tokens.
		 */
		{
			MatcherCode(StringToken),
			CompileMatcher(`"[^"]*"`),
		},

		{
			MatcherCode(ColonToken),
			CompileMatcher(`:`),
		},

		{
			MatcherCode(CommaToken),
			CompileMatcher(`,`),
		},

		/*
		 * The "else if" token is parsed before the "else", "if", and space tokens because it
		 * contains both.
		 */
		{
			MatcherCode(ElseIfToken),
			CompileMatcher("else if"),
		},

		{
			MatcherCode(ElseToken),
			CompileMatcher("else"),
		},

		{
			MatcherCode(IfToken),
			CompileMatcher("if"),
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

		/*
		 * Operators and identifiers are parsed last because they shouldn't contain anything that
		 * would be _identified_ (get it?) as another token.
		 *
		 * Operators can contain any special (non-control, non-alphanumeric) ASCII character with
		 * the following exceptions.
		 *
		 * Conflicts with other tokens:
		 * "\"", "(", ")", ",", ".", ":", "_"
		 *
		 * Reserved for future use:
		 * "#", "$", ",", "/", ";", "?", "@", "[", "]", "\\", "`", "{", "}"
		 *
		 * "=" is also not a valid operator.
		 */
		{
			MatcherCode(OperatorToken),
			CompileMatcher(`[!%&*+\-<=>^|~]*=[!%&*+\-<=>^|~]+|[!%&*+\-<=>^|~]+=[!%&*+\-<=>^|~]*|[!%&*+\-<>^|~]+`),
		},

		/*
		 * The assignment operator token is parsed after the operator token so an operator
		 * containing multiple adjacent equals symbols isn't parsed as multiple assignment operator
		 * tokens.
		 */
		{
			MatcherCode(AssignmentOperatorToken),
			CompileMatcher(`=`),
		},

		{
			MatcherCode(IdentifierToken),
			CompileMatcher(`[^\t\n !"%&()*+,\-.<=>]+`),
		},
	},
}

type Lexer struct {
	cachedTokens []*lexer.Token
	fileContent  string
	filename     string
	i            int
}

func indentCharacterAndCount(line string) (rune, int) {
	lineCharacters := []rune(line)

	if len(lineCharacters) == 0 || (lineCharacters[0] != '\t' && lineCharacters[0] != ' ') {
		return 0, 0
	}

	indentCharacter := lineCharacters[0]
	indentCount := 0

	for indentCount < len(lineCharacters) && lineCharacters[indentCount] == indentCharacter {
		indentCount++
	}

	return indentCharacter, indentCount
}

func isLineBlank(line string) bool {
	for _, character := range line {
		if character != '\t' && character != ' ' {
			return false
		}
	}

	return true
}

func (lexer_ *Lexer) Next() (lexer.Token, error) {
	if lexer_.cachedTokens == nil {
		lexer_.cachedTokens = lexer_.tokens()

		if lexer_.cachedTokens == nil {
			errors.RaiseError(lexer_errors.LexerFailed)
		}
	}

	if lexer_.i == len(lexer_.cachedTokens) {
		return lexer.EOFToken(
			lexer.Position{
				Filename: lexer_.filename,
				Offset:   len(lexer_.fileContent),

				/*
				 * These fields aren't populated because we don't use them.
				 *
				 * Generally, they're useful only for error handling, but we already reconstruct the
				 * line and column when we raise errors (see
				 * `src/interpreter/errors/code_highlighter.go` to understand how).
				 *
				 * See `src/interpreter/main.go`, where we use tokens' positions to raise errors.
				 */
				Line:   0,
				Column: 0,
			},
		), nil
	}

	token := lexer_.cachedTokens[lexer_.i]

	lexer_.i++

	return *token, nil
}

func (lexer_ *Lexer) tokens() []*lexer.Token {
	if len(lexer_.fileContent) == 0 {
		return []*lexer.Token{}
	}

	matches := matcher.MatchWithInitial(MatcherInput(lexer_.fileContent), lexer_.parseIndentation())

	if matches == nil {
		return nil
	}

	result := []*lexer.Token{}

	for _, match := range matches {
		if match.Type != MatcherCode(SpaceToken) {
			result = append(result, &lexer.Token{
				Type:  lexer.TokenType(match.Type),
				Value: lexer_.fileContent[match.Start:match.End],
				Pos: lexer.Position{
					Filename: lexer_.filename,
					Offset:   match.Start,

					// Similarly, these fields aren't populated because we don't use them either.
					Line:   0,
					Column: 0,
				},
			})
		}
	}

	return result
}

/*
 * Before we tokenize the input, we replace indentation with indent and outdent meta-tokens, which
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
func (lexer_ *Lexer) parseIndentation() []*ExhaustiveMatch {
	result := []*ExhaustiveMatch{}
	addMatchToResult := func(type_ MatcherCode, start int, end int) {
		result = append(result, &ExhaustiveMatch{
			Type:      type_,
			Start:     start,
			End:       end,
			Subgroups: [][2]int{},
		})
	}

	endOfLastMatch := func() int {
		if len(result) == 0 {
			return 0
		}

		return result[len(result)-1].End
	}

	fileOffset := 0
	indentCharacter := rune(0)
	indentLength := 0
	lastIndentCount := 0

	for _, line := range strings.Split(lexer_.fileContent, "\n") {
		if !isLineBlank(line) {
			currentIndentCharacter, indentCount := indentCharacterAndCount(line)

			if indentCharacter == 0 {
				indentCharacter = currentIndentCharacter
				indentLength = indentCount
			}

			if currentIndentCharacter != 0 &&
				(currentIndentCharacter != indentCharacter ||
					(indentLength > 0 && indentCount%indentLength > 0)) {
				errors.RaisePositionalError(
					&errors.PositionalError{
						Error: lexer_errors.InconsistentIndentation(
							indentCharacter,
							indentLength,
							currentIndentCharacter,
							indentCount,
						),

						Position: &errors.Position{
							Start: fileOffset,
							End:   fileOffset + indentCount,
						},
					},

					lexer_.fileContent,
				)
			}

			if indentLength > 0 {
				indentCount /= indentLength
			}

			lineStart := fileOffset + indentCount*indentLength

			if indentCount < lastIndentCount {
				for i := 0; i < lastIndentCount-indentCount; i++ {
					addMatchToResult(
						MatcherCode(OutdentToken),
						endOfLastMatch(),
						endOfLastMatch(),
					)
				}
			}

			if endOfLastMatch() > 0 {
				addMatchToResult(UnrecognizedMatcherCode, endOfLastMatch(), fileOffset)
			}

			if indentCount > lastIndentCount {
				for i := 0; i < indentCount-lastIndentCount; i++ {
					addMatchToResult(MatcherCode(IndentToken), lineStart, lineStart)
				}
			}

			addMatchToResult(UnrecognizedMatcherCode, lineStart, fileOffset+len(line))

			lastIndentCount = indentCount
		}

		fileOffset += len(line) + 1
	}

	for i := 0; i < lastIndentCount; i++ {
		addMatchToResult(MatcherCode(OutdentToken), endOfLastMatch(), endOfLastMatch())
	}

	if endOfLastMatch() < len(lexer_.fileContent) {
		addMatchToResult(UnrecognizedMatcherCode, endOfLastMatch(), len(lexer_.fileContent))
	}

	return result
}

type LexerDefinition struct{}

func (definition *LexerDefinition) Lex(filename string, reader io.Reader) (lexer.Lexer, error) {
	fileContent, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return definition.LexString(filename, string(fileContent))
}

func (definition *LexerDefinition) LexString(filename string, fileContent string) (lexer.Lexer, error) {
	return &Lexer{
		cachedTokens: nil,
		fileContent:  fileContent,
		filename:     filename,
		i:            0,
	}, nil
}

func (definition *LexerDefinition) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"AssignmentOperatorToken": lexer.TokenType(AssignmentOperatorToken),
		"ColonToken":              lexer.TokenType(ColonToken),
		"CommaToken":              lexer.TokenType(CommaToken),
		"FloatToken":              lexer.TokenType(FloatToken),
		"FunctionKeywordToken":    lexer.TokenType(FunctionKeywordToken),
		"IdentifierToken":         lexer.TokenType(IdentifierToken),
		"IndentToken":             lexer.TokenType(IndentToken),
		"OutdentToken":            lexer.TokenType(OutdentToken),
		"IntegerToken":            lexer.TokenType(IntegerToken),
		"LeftParenthesisToken":    lexer.TokenType(LeftParenthesisToken),
		"RightParenthesisToken":   lexer.TokenType(RightParenthesisToken),
		"NewlineToken":            lexer.TokenType(NewlineToken),
		"OperatorToken":           lexer.TokenType(OperatorToken),
		"SelectOperatorToken":     lexer.TokenType(SelectOperatorToken),
		"StringToken":             lexer.TokenType(StringToken),
		"EOF":                     lexer.EOF,
	}
}
