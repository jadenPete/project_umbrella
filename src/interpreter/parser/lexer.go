package parser

type TokenType MatcherCode
type Token struct {
	type_ TokenType
	start int
	end   int
}

const (
	StringToken TokenType = iota + 1
	AdditionOperatorToken
	AssignmentOperatorToken
	CommaToken
	IdentifierToken
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
			MatcherCode(AdditionOperatorToken),
			CompileMatcher(`\+`),
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
			MatcherCode(SpaceToken),
			CompileMatcher(`[\t ]+`),
		},

		{
			MatcherCode(SelectOperatorToken),
			CompileMatcher(`\.`),
		},

		{
			MatcherCode(IdentifierToken),
			CompileMatcher(`[^\n\t "=,()]+`),
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
