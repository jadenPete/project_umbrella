package main

import "github.com/alecthomas/participle/v2"

func main() {
	parser := participle.MustBuild[Expression](
		participle.Lexer(&LexerDefinition{}),
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
