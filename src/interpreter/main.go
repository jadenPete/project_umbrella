package main

import (
	"os"

	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/entry_errors"
	"project_umbrella/interpreter/errors/parser_errors"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/runtime"
)

func executeSource(source string) {
	concreteTree, err := parser.ParseString(source)

	if err != nil {
		var participleError participle.Error
		var participleErrorPosition *errors.Position

		switch err := err.(type) {
		case *participle.ParseError:
			participleError = err
			participleErrorPosition = &errors.Position{
				Start: err.Pos.Offset,
				End:   err.Pos.Offset + 1,
			}

		case *participle.UnexpectedTokenError:
			participleError = err
			participleErrorPosition = &errors.Position{
				Start: err.Unexpected.Pos.Offset,
				End:   err.Unexpected.Pos.Offset + len(err.Unexpected.Value),
			}

		default:
			panic(err)
		}

		errors.RaisePositionalError(
			&errors.PositionalError{
				Error:    parser_errors.ParserFailed(participleError),
				Position: participleErrorPosition,
			},

			source,
		)
	}

	abstractTree := concreteTree.Abstract()

	runtime.ExecuteBytecode(bytecode_generator.ExpressionToBytecodeFromCache(abstractTree, source))
}

func main() {
	if len(os.Args) < 2 {
		errors.RaiseError(entry_errors.FileNotSpecified)
	}

	fileContent, err := os.ReadFile(os.Args[1])

	if err != nil {
		errors.RaiseError(entry_errors.FileNotOpened)
	}

	executeSource(string(fileContent))
}
