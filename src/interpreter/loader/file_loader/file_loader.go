package file_loader

import (
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"

	"project_umbrella/interpreter/bytecode_generator"
	"project_umbrella/interpreter/common"
	"project_umbrella/interpreter/environment_variables"
	"project_umbrella/interpreter/errors"
	"project_umbrella/interpreter/errors/entry_errors"
	"project_umbrella/interpreter/errors/parser_errors"
	"project_umbrella/interpreter/loader"
	"project_umbrella/interpreter/parser"
	"project_umbrella/interpreter/runtime/runtime_executor"
	"project_umbrella/interpreter/runtime/value"
)

func expressionListFromSource(
	path string,
	source string,
	loaderChannel *loader.LoaderChannel,
) *parser.ExpressionList {
	concreteResult, err := parser.ParseString(path, source)

	if err != nil {
		var participleError participle.Error
		var participleErrorPosition *errors.Position

		switch err := err.(type) {
		case *participle.ParseError:
			participleError = err
			participleErrorPosition = &errors.Position{
				Filename: path,
				Start:    err.Pos.Offset,
				End:      err.Pos.Offset + 1,
			}

		case *participle.UnexpectedTokenError:
			participleError = err
			participleErrorPosition = &errors.Position{
				Filename: path,
				Start:    err.Unexpected.Pos.Offset,
				End:      err.Unexpected.Pos.Offset + len(err.Unexpected.Value),
			}

		default:
			panic(err)
		}

		errors.RaisePositionalError(
			&errors.PositionalError{
				Error:    parser_errors.ParserFailed(participleError),
				Position: participleErrorPosition,
			},
		)
	}

	return concreteResult.AbstractExpressionList()
}

func expressionListFromStartupFile(
	sourcePath string,
	loaderChannel *loader.LoaderChannel,
) *parser.ExpressionList {
	excludedDirectories := strings.Split(environment_variables.KRAIT_STARTUP_EXCLUDE, ":")
	emptyResult := &parser.ExpressionList{
		Children_: []parser.Expression{},
	}

	for _, excludedDirectory := range excludedDirectories {
		if excludedDirectory != "" &&
			common.IsDirectoryAncestorOfFile(excludedDirectory, sourcePath) {
			return &parser.ExpressionList{
				Children_: []parser.Expression{},
			}
		}
	}

	if environment_variables.KRAIT_STARTUP == "" {
		return emptyResult
	}

	startupFileContent, err := os.ReadFile(environment_variables.KRAIT_STARTUP)

	if err != nil {
		errors.RaiseError(entry_errors.StartupFileNotOpened(environment_variables.KRAIT_STARTUP))
	}

	return expressionListFromSource(
		environment_variables.KRAIT_STARTUP,
		string(startupFileContent),
		loaderChannel,
	)
}

func LoadFile(path string, loaderChannel *loader.LoaderChannel) value.Value {
	fileContentByteSlice, err := os.ReadFile(path)

	if err != nil {
		errors.RaiseError(entry_errors.FileNotOpened(path))
	}

	fileContent := string(fileContentByteSlice)
	expressionList := (&parser.ExpressionList{
		Children_: append(
			expressionListFromStartupFile(path, loaderChannel).Children_,
			expressionListFromSource(path, fileContent, loaderChannel).Children_...,
		),
	}).ToModule()

	return runtime_executor.ExecuteBytecode(
		bytecode_generator.ExpressionToBytecodeFromCache(expressionList, fileContent),
		loaderChannel,
	)
}
