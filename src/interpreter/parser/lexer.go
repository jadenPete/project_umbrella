package parser

type TokenType MatcherCode
type Token struct {
	type_ TokenType
	start int
	end   int
}

const (
	StringToken TokenType = iota + 1
	AssignmentOperatorToken
	CommaToken
	FloatToken
	IdentifierToken
	IntegerToken
	LeftParenthesisToken
	RightParenthesisToken
	NewlineToken
	SpaceToken
	SelectOperatorToken
)

var matcher = ExhaustiveMatcher{
	patterns: []*ExhaustiveMatchPattern{
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
			CompileMatcher(`\n+`),
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

/*
 * Tokenize the input file content, converting it into a slice of tokens.
 *
 * If tokenization failed, this function returns `nil`.
 */
func (lexer *Lexer) Parse() []*Token {
	if len(lexer.FileContent) == 0 {
		return make([]*Token, 0)
	}

	matches := matcher.Match(MatcherInput(lexer.FileContent))

	if matches == nil {
		return nil
	}

	result := make([]*Token, 0)

	for _, match := range matches {
		if match.type_ != MatcherCode(SpaceToken) {
			result = append(result, &Token{
				type_: TokenType(match.type_),
				start: match.start,
				end:   match.end,
			})
		}
	}

	return result
}
