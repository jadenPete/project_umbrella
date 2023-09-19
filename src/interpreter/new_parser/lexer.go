package main

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"

	"project_umbrella/interpreter/parser"
)

type Lexer struct {
	filename  string
	oldLexer  *parser.Lexer
	oldTokens []*parser.Token
	i         int
}

func (lexer_ *Lexer) Next() (lexer.Token, error) {
	if lexer_.oldTokens == nil {
		lexer_.oldTokens = lexer_.oldLexer.Parse()
	}

	if lexer_.i == len(lexer_.oldTokens) {
		return lexer.EOFToken(
			lexer.Position{
				Filename: lexer_.filename,
				Offset:   len(lexer_.oldLexer.FileContent),

				// TODO: Compute these fields' values
				Line:   0,
				Column: 0,
			},
		), nil
	}

	token := lexer_.oldTokens[lexer_.i]

	lexer_.i++

	return lexer.Token{
		Type:  lexer.TokenType(token.Type),
		Value: lexer_.oldLexer.FileContent[token.Position.Start:token.Position.End],
		Pos: lexer.Position{
			Filename: lexer_.filename,
			Offset:   token.Position.Start,

			// TODO: Compute these fields' values
			Line:   0,
			Column: 0,
		},
	}, nil
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
	oldLexer := &parser.Lexer{
		FileContent: fileContent,
	}

	return &Lexer{
		filename:  filename,
		oldLexer:  oldLexer,
		oldTokens: nil,
		i:         0,
	}, nil
}

func (definition *LexerDefinition) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"AssignmentOperatorToken": lexer.TokenType(parser.AssignmentOperatorToken),
		"ColonToken":              lexer.TokenType(parser.ColonToken),
		"CommaToken":              lexer.TokenType(parser.CommaToken),
		"FloatToken":              lexer.TokenType(parser.FloatToken),
		"FunctionKeywordToken":    lexer.TokenType(parser.FunctionKeywordToken),
		"IdentifierToken":         lexer.TokenType(parser.IdentifierToken),
		"IndentToken":             lexer.TokenType(parser.IndentToken),
		"OutdentToken":            lexer.TokenType(parser.OutdentToken),
		"IntegerToken":            lexer.TokenType(parser.IntegerToken),
		"LeftParenthesisToken":    lexer.TokenType(parser.LeftParenthesisToken),
		"RightParenthesisToken":   lexer.TokenType(parser.RightParenthesisToken),
		"NewlineToken":            lexer.TokenType(parser.NewlineToken),
		"SelectOperatorToken":     lexer.TokenType(parser.SelectOperatorToken),
		"StringToken":             lexer.TokenType(parser.StringToken),
		"EOF":                     lexer.EOF,
	}
}
