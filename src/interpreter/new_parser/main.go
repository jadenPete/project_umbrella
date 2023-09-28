package main

import (
	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/parser"
)

func main() {
	parser := participle.MustBuild[ConcreteExpression](
		participle.Lexer(&parser.LexerDefinition{}),
		participle.Union[ConcreteStatement](&ConcreteAssignment{}, &ConcreteFunction{}, &ConcreteInfixAddition{}),
		participle.Union[ConcreteExpression](&ConcreteInfixAddition{}),
		participle.Union[ConcretePrimary](&ConcreteFloat{}, &ConcreteIdentifier{}, &ConcreteInteger{}, &ConcreteString{}),
	)

	tree, err := parser.ParseString("", `2 + println(2 + 2) + 2`)

	if err != nil {
		println(err.Error())
	}

	println(tree)
}
