package main

import (
	"os"

	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/entry_errors"
	"project_umbrella/interpreter/errors/lexer_errors"
	"project_umbrella/interpreter/errors/parser_errors"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/runtime"
)

func executeSource(source string) {
	lexer := &parser.Lexer{
		FileContent: source,
	}

	tokens := lexer.Parse()

	if tokens == nil {
		errors.RaiseError(lexer_errors.LexerFailed)
	}

	expression := (&parser.Parser{
		FileContent: source,
		Tokens:      tokens,
	}).Parse()

	if expression == nil {
		errors.RaiseError(parser_errors.ParserFailed)
	}

	runtime.ExecuteBytecode(parser.ExpressionToBytecodeFromCache(expression, source))
}

func main() {
	if len(os.Args) < 2 {
		errors.RaiseError(entry_errors.FileNotSpecified)
	}

	content, err := os.ReadFile(os.Args[1])

	if err != nil {
		errors.RaiseError(entry_errors.FileNotOpened)
	}

	executeSource(string(content))
}
