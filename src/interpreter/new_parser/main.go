package main

import (
	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/parser"
)

func main() {
	parser := participle.MustBuild[Expression](
		participle.Lexer(&parser.LexerDefinition{}),
		participle.Union[Statement](&Assignment{}, &Function{}, &InfixAddition{}),
		participle.Union[Expression](&InfixAddition{}),
		participle.Union[Primary](&Float{}, &Identifier{}, &Integer{}, &String{}),
	)

	tree, err := parser.ParseString("", `2 + println(2 + 2) + 2`)

	if err != nil {
		println(err.Error())
	}

	println(tree)
}
