package main

import (
	"os"

	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/runtime"
)

func executeSource(source string) {
	lexer := &parser.Lexer{
		FileContent: source,
	}

	tokens := lexer.Parse()

	if tokens == nil {
		panic("The lexer failed.")
	}

	expression := (&parser.Parser{
		FileContent: source,
		Tokens:      tokens,
	}).Parse()

	if expression == nil {
		panic("The parser failed.")
	}

	runtime.ExecuteBytecode(parser.ExpressionToBytecodeFromCache(expression, source))
}

func main() {
	if len(os.Args) < 2 {
		panic("Please specify a file to run.")
	}

	content, err := os.ReadFile(os.Args[1])

	if err != nil {
		panic("Couldn't open the specified file.")
	}

	executeSource(string(content))
}
